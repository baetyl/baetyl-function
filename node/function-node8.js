#!/usr/bin/env node
const path = require('path');
const fs = require('fs');
const log4js = require('log4js');
const moment = require('moment');
const services = require('./function_grpc_pb.js');
const grpc = require('grpc');
const YAML = require('yaml');
const util = require('util');
const errFormat = 'Exception calling application: %s'
const address = '0.0.0.0:50050'
const confPath = path.join('etc', 'baetyl', 'service.yml')
const codePath = path.join('var', 'lib', 'baetyl', 'code')

const hasAttr = (obj, attr) => {
    if (obj instanceof Object && !(obj instanceof Array)) {
        if (obj[attr] != undefined) {
            return true;
        }
    }
    return false;
};

const getLogger = config => {
    if (!hasAttr(config, 'logger')) {
        return log4js.getLogger(config.name);
    }
    if (!hasAttr(config.logger, 'path')) {
        return log4js.getLogger(config.name);
    }
    let level = 'info';
    if (hasAttr(config.logger, 'level')) {
        level = config.logger.level;
    }

    let backupCount = 15;
    if (hasAttr(config.logger, 'backupCount') && hasAttr(config.logger.backupCount, 'max')) {
        backupCount = config.logger.backupCount.max;
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
                filename: config.logger.path,
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
    const logger = log4js.getLogger(config.name);
    return logger;
};

const LoadFunctions = (file, functionsHandle) => {
    config = YAML.parse(fs.readFileSync(file).toString());
    if (!hasAttr(config, 'functions')) {
         throw new Error('Module config invalid, missing functions');
    }
    config.functions.forEach(function (ele) {
        if (ele.name == undefined || ele.handler == undefined || ele.codedir == undefined) {
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

const getFunctions = () => {
    functionsHandle = {};
    var list = fs.readdirSync(codePath)
    list.forEach(function(file) {
        if (path.extname(file)=='.yml') {
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
    Load(confpath) {
        this.config = {}
        if (fs.existsSync(confpath)) {
            this.config = YAML.parse(fs.readFileSync(confpath).toString());
        }
        if (hasAttr(process.env, 'SERVICE_NAME')) {
            this.config['name'] = process.env['SERVICE_NAME']
        } 

        this.config['server'] = {
            'address': address
        }
        
        if (!hasAttr(this.config, 'name')) {
            throw new Error('Module config invalid, missing name');
        }
        if (!hasAttr(this.config, 'server')) {
            throw new Error('Module config invalid, missing server');
        }
        if (!hasAttr(this.config.server, 'address')) {
            throw new Error('Module config invalid, missing server address');
        }

        // Need Logger here ?
        this.logger = getLogger(this.config);
        const functionsHandle = getFunctions();
        this.server = getGrpcServer(this.config);

        this.server.addService(services.FunctionService, {
            invoke: (invoke, callback) => (this.Invoke(functionsHandle, invoke, callback))
        });
    }
    Start() {
        // grpc server
        this.server.start();
        this.logger.info('service starting');
    }
    Close(callback) {
        if (hasAttr(this.config.server, 'timeout')) {
            const timeout = new Number(this.config.server.timeout / 1e9);
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
    Invoke(functionsHandle, invoke, callback) {
        // switch (invoke.request.getType()) {
        //     case "HTTP":

        // }
        // assume this is HTTP
        const ctx = {};
        ctx.invokeId = invoke.getMetadata()['invokeId']
        ctx.functionName = invoke.getService()

        event = {}
        event.resource = invoke.request.getMetadata()['resource']
        event.path = invoke.request.getMetadata()['path']
        event.httpMethod = invoke.request.getMetadata()['httpMethod']
        //TODO: 判空？
        event.headers = json.parse(invoke.request.getMetadata()['headers'])
        event.queryStringParameters = json.parse(invoke.request.getMetadata()['queryStringParameters'])
        event.requestContext = json.parse(invoke.request.getMetadata()['requestContext'])
        //TODO: 判空？
        event.headers = json.parse(invoke.request.getMetadata()['headers'])
        // TO string
        event.body = string(invoke.request.getPayload())
        event.isBase64Encoded = invoke.request.getMetadata()['body']

        var method = invoke.request.getMethod()
        if (method == "") {
            method = functionsHandle[0]
        }
        if (method == undefined) {
            return callback(new Error(util.format(errFormat, "function not found")));
        }
        functionHandle(
            msg,
            ctx,
            (err, respMsg) => {
                if (err != null) {
                    err = util.format('(\'[UserCodeInvoke]\', %s)', err.toString())
                    return callback(new Error(util.format(errFormat, err)));
                }

                if (respMsg == "" || respMsg == undefined || respMsg == null) {
                    call.request.setPayload("");
                } else if (Buffer.isBuffer(respMsg)) {
                    call.request.setPayload(respMsg);
                }
                else {
                    try {
                        const jsonString = JSON.stringify(respMsg);
                        call.request.setPayload(Buffer.from(jsonString));
                    }
                    catch (error) {
                        err = util.format('(\'[UserCodeReturn]\', %s)', error.toString())
                        return callback(new Error(util.format(errFormat, err)));
                    }
                }
                callback(null, call.request);
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
