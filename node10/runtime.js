#!/usr/bin/env node

const path = require('path');
const fs = require('fs');
const log4js = require('log4js');
const moment = require('moment');
const grpc = require('grpc');
const YAML = require('yaml');
const querystring = require('querystring');
const services = require('./function_grpc_pb.js');
const messages = require('./function_pb.js');
const _HeaderDelim = '&__header_delim__&'
const _HeaderEquals = '&__header_equals__&'


const hasAttr = (obj, attr) => {
    if (obj instanceof Object && !(obj instanceof Array)) {
        if (obj[attr] != undefined) {
            return true;
        }
    }
    return false;
};

const parseHttpParams = request => {
    let event = {}
    event['path'] = request.getMetadataMap().get('path')
    event['resource'] = event['path']
    event['httpMethod'] = request.getMetadataMap().get('httpMethod')
    event['pathParameters'] = {}
    event['body'] = request.getPayload()
    event['isBase64Encoded'] = request.getMetadataMap().get('isBase64Encoded')
    event['queryStringParameters'] = querystring.parse(
        request.getMetadataMap().get('queryStringParameters'))
    event['headers'] = request.getMetadataMap().get('headers').split(_HeaderDelim)
    let headers = {}
    event['headers'].forEach(function(header){
        let kv = header.split(_HeaderEquals)
        headers[kv[0]] = kv[1]
    });
    event['headers'] = headers
    event['requestContext'] = {
        "stage": "",
        "requestId": request.getMetadataMap().get('invokeId'),
        "resourcePath": event['resource'],
        "httpMethod": event['httpMethod'],
        "apiId": "",
        "sourceIp": "",
    }
    return event
}

const populateHttpResponse = (callback, code, msg) => {
    let message = new messages.Message()
    message.getMetadataMap().set('statusCode', code.toString())

    let payload = {
        "errorCode": code.toString(),
        "message": msg,
    };
    message.setPayload(Buffer.from(JSON.stringify(payload)).toString('base64'));
    return callback(null, message)
}

const getLogger = mo => {
    if (!hasAttr(mo, 'logger') || !hasAttr(mo.config.logger, 'path')) {
        let logger = log4js.getLogger(mo.name);
        logger.level = 'info'
        return logger;
    }
   
    let level = 'info';
    if (hasAttr(mo.config.logger, 'level')) {
        level = mo.config.logger.level;
    }

    let backupCount = 15;
    if (hasAttr(mo.config.logger, 'backupCount') && hasAttr(mo.config.logger.backupCount, 'max')) {
        backupCount = mo.config.logger.backupCount.max;
    }
    log4js.addLayout('baetyl', () => logEvent => {
        const asctime = moment(logEvent.startTime).format('YYYY-MM-DD HH:mm:ss');
        const name = logEvent.categoryName;
        const levelname = logEvent.level.levelStr;
        const message = logEvent.data;
        return `${asctime} - ${name} - ${levelname} - ${message}`;
    });
    log4js.configure({
        appenders: {
            file: {
                type: 'file',
                filename: mo.config.logger.path,
                layout: { type: 'baetyl' },
                backups: backupCount,
                compress: true,
                encoding: 'utf-8'
            }
        },
        categories: {
            default: { appenders: ['file'], level }
        }
    });
    return log4js.getLogger(mo.config.name)
};

const LoadFunctions = (file, functionsHandle) => {
    let config = YAML.parse(fs.readFileSync(file).toString());
    if (!hasAttr(config, 'functions')) {
         throw new Error('Module config invalid, missing functions');
    }
    config.functions.forEach(function (ele) {
        if (ele.name === undefined || ele.handler === undefined || ele.codedir === undefined) {
            throw new Error('config invalid, missing function name, handler or codedir');
        }
        const codedir = ele.codedir;
        const moduleHandler = ele.handler.split('.');
        const handlerName = moduleHandler[1];
        const moduleName = require(path.join(path.dirname(file), codedir, moduleHandler[0]));
        const functionHandle = moduleName[handlerName];
        functionsHandle[ele.name] = functionHandle;
    });
}

const getFunctions = codePath => {
    let functionsHandle = {};
    codePath = path.join(process.cwd(), codePath)

    let list = fs.readdirSync(codePath)
    list.forEach(function(file) {
        if (path.extname(file)==='.yml') {
            LoadFunctions(path.join(codePath, file), functionsHandle)
        }
    })
    return functionsHandle;
};

const getGrpcServer = config => {
    let maxMessageSize = 4 * 1024 * 1024;
    if (hasAttr(config['server'], 'message')
        && hasAttr(config['server']['message'], 'length')
        && hasAttr(config['server']['message']['length'], 'max')) {
        maxMessageSize = config['server']['message']['length']['max'];
    }
    let server = new grpc.Server({
        'grpc.max_send_message_length': maxMessageSize,
        'grpc.max_receive_message_length': maxMessageSize
    });

    let credentials = undefined;

    if (hasAttr(config.server, 'ca')
        && hasAttr(config.server, 'key')
        && hasAttr(config.server, 'cert')) {

        credentials = grpc.ServerCredentials.createSsl(
            fs.readFileSync(config['server']['ca']), [{
                cert_chain: fs.readFileSync(config['server']['cert']),
                private_key: fs.readFileSync(config['server']['key'])
            }], true);
    } else {
        credentials = grpc.ServerCredentials.createInsecure();
    }

    server.bind(config['server']['address'], credentials);
    return server;
}

class NodeRuntimeModule {
    constructor() {
        this.name = 'baetyl-function'
        this.confPath = '/etc/baetyl/service.yml'
        this.codePath = '/var/lib/baetyl/code'
        // this.codePath = '../testdata/node10'
        this.serverAddress = "0.0.0.0:80"
    }

    Load() {
        if (hasAttr(process.env, 'SERVICE_NAME')) {
            this.name = process.env['SERVICE_NAME']
        } 

        if (hasAttr(process.env, 'SERVICE_CONF')) {
            this.confPath = process.env['SERVICE_CONF']
        }

        if (hasAttr(process.env, 'SERVICE_CODE')) {
            this.codePath = process.env['SERVICE_CODE']
        }

        if (hasAttr(process.env, 'SERVICE_ADDRESS')) {
            this.serverAddress = process.env['SERVICE_ADDRESS']
        }

        this.config = {
            'server': {
                'address': this.serverAddress
            }
        }

        if (fs.existsSync(this.confPath)) {
            this.config = YAML.parse(fs.readFileSync(this.confPath).toString());
        }
        
        this.logger = getLogger(this);
        this.functionsHandle = getFunctions(this.codePath);
        this.server = getGrpcServer(this.config);

        this.server.addService(services.FunctionService, {
            call: (call, callback) => (this.Call(call, callback))
        });
    }
    Start() {
        // grpc server
        this.logger.info('service starting');
        this.server.start();
    }
    Close(callback) {
        if (hasAttr(this.config.server, 'timeout')) {
            const timeout = Number(this.config.server.timeout / 1e9);
            setTimeout(() => {
                this.server.forceShutdown();
                this.logger.info('service closed');
                callback();
            }, timeout);
        } else {
            this.server.forceShutdown();
            this.logger.info('service closed');
        }
    }
    Call(call, callback) {
        switch (call.request.getType()) {
            case "HTTP":
                this.parseHttp(call, callback)
                break

            default:
                callback(new Error("Type of Message doesn't support"), null)
        }
    }
    parseHttp(call, callback) {
        let event = parseHttpParams(call.request)

        let ctx = {
            'invokeid': call.request.getMetadataMap().get('invokeId'),
            'functionName': call.request.getName()
        }

        let method = call.request.getMethod()
        if (method === "") {
            method = Object.keys(this.functionsHandle)[0]
        }

        if (!hasAttr(this.functionsHandle, method)) {
            this.logger.info("no route to method: %s", method)
            return populateHttpResponse(callback, 404, 'no route');
        }

        let functionHandle = this.functionsHandle[method];
        functionHandle(
            event,
            ctx,
            (err, respMsg) => {
                if (err != null) {
                    this.logger.info("error when executing method %s: %s", method, err)
                    return populateHttpResponse(callback, 500, err);
                }

                let type = typeof(respMsg)
                if (!(type === 'object' || type === 'string')) {
                    this.logger.info("function response error: %s",
                          "response is not object or string")
                    return populateHttpResponse(callback, 502, "function response error")
                }

                if (type === 'string') {
                    try {
                        respMsg = JSON.parse(respMsg)
                    } catch(ex) {
                        this.logger.info(
                            "function response error in loads response: %s", ex)
                        return populateHttpResponse(callback, 502, "function response error")
                    }
                }

                let message = new messages.Message()

                if (!hasAttr(respMsg, 'statusCode') || typeof(respMsg['statusCode']) != 'number') {
                    this.logger.info("function response error: %s", "missing statusCode")
                    return populateHttpResponse(callback, 502, "function response error")
                }else{
                    message.getMetadataMap().set('statusCode', respMsg['statusCode'].toString())
                }

                if (hasAttr(respMsg, 'headers')) {
                    if (typeof(respMsg['headers']) != 'object') {
                        this.logger.info("function response error: %s",
                              "headers is not dict")
                        return populateHttpResponse(callback, 502, "function response error")
                    }
                } else{
                    respMsg['headers'] = {}
                }

                if (hasAttr(respMsg, 'isBase64Encoded')) {
                    if (typeof(respMsg['isBase64Encoded']) === "boolean") {
                        message.getMetadataMap().set('isBase64Encoded', respMsg['isBase64Encoded'].toString())
                    } else {
                        this.logger.info("function response error: %s",
                              "isBase64Encoded is not bool")
                        return populateHttpResponse(callback, 502, 'function response error')
                    }
                }

                if (hasAttr(respMsg, 'body')) {
                    if(typeof(respMsg['body']) == 'string'){
                        message.setPayload(Buffer.from(respMsg['body']).toString('base64'))
                    } else {
                        this.logger.info("function response error: %s", "body is not str")
                        return populateHttpResponse(callback, 502, "function response error")
                    }

                    try {
                        JSON.parse(respMsg['body'])
                        respMsg['headers']['Content-Type'] = 'application/json'
                    } catch(ex) {
                        respMsg['headers']['Content-Type'] = 'text/plain'
                    }

                    let items = []
                    for (let k in respMsg['headers']) {
                        if (typeof(respMsg['headers'][k]) != 'string') {
                            this.logger.info("function response error: %s",
                                  "value in headers is not str");
                            return populateHttpResponse(callback, 502, "function response error")
                        } 
                        items.push(k + _HeaderEquals + respMsg['headers'][k]);
                    }
                    message.getMetadataMap().set('headers', items.join(_HeaderDelim));
                    message.setPayload(Buffer.from(respMsg['body']).toString('base64'));
                }
                callback(null, message);
            }
        );
    }
}

(() => {
    const runtimeModule = new NodeRuntimeModule();
    runtimeModule.Load();
    runtimeModule.Start();
    function closeServer() {
        runtimeModule.Close(() => log4js.shutdown(() => process.exit(0)));
    }
    process.on('SIGINT', () => {
        closeServer();
    });
    process.on('SIGTERM', () => {
        closeServer();
    });
})();
