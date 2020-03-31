#!/usr/bin/env node

const path = require('path');
const fs = require('fs');
const log4js = require('log4js');
const moment = require('moment');
const grpc = require('grpc');
const yaml = require('yaml');
const services = require('./function_grpc_pb.js');


const hasAttr = (obj, attr) => {
    if (obj instanceof Object && !(obj instanceof Array)) {
        if (obj[attr] !== undefined) {
            return true;
        }
    }
    return false;
};

const getLogger = mo => {
    if (!hasAttr(mo, 'logger') || !hasAttr(mo.config.logger, 'path')) {
        let logger = log4js.getLogger(mo.name);
        logger.level = 'info';
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
    let config = yaml.parse(fs.readFileSync(file).toString());
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
        functionsHandle[ele.name] = moduleName[handlerName];
    });
};

const getFunctions = codePath => {
    let functionsHandle = {};
    codePath = path.join(process.cwd(), codePath);
    if (!fs.existsSync(codePath)) {
        throw new Error('no such file or directory: ' + codePath);
    }

    let list = fs.readdirSync(codePath);
    list.forEach(function(file) {
        if (path.extname(file)==='.yml') {
            LoadFunctions(path.join(codePath, file), functionsHandle)
        }
    });
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
};

class NodeRuntimeModule {
    constructor() {
        this.name = 'baetyl-node10';
        this.confPath = '/etc/baetyl/service.yml';
        this.codePath = '/var/lib/baetyl/code';
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
        };

        if (fs.existsSync(this.confPath)) {
            this.config = yaml.parse(fs.readFileSync(this.confPath).toString());
        }
        
        this.logger = getLogger(this);
        this.functionsHandle = getFunctions(this.codePath);
        this.server = getGrpcServer(this.config);

        this.server.addService(services.FunctionService, {
            call: (call, callback) => (this.Call(call, callback))
        });
    }
    Start() {
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
        let functionName = call.request.getMetadataMap().get('functionName');
        if (functionName === "") {
            functionName = Object.keys(this.functionsHandle)[0]
        }

        if (!hasAttr(this.functionsHandle, functionName)) {
            this.logger.error("the function doesn't found: %s", functionName);
            return callback(new Error("the function doesn't found"));
        }

        let ctx = {};
        call.request.getMetadataMap().forEach(function (v, k) {
            ctx[k] = v
        });

        let msg = '';
        const Payload = call.request.getPayload();
        if (Payload) {
            try {
                const payloadString = Buffer.from(Payload).toString();
                msg = JSON.parse(payloadString);
            }
            catch (error) {
                msg = Buffer.from(Payload); // raw data, not json format
            }
        }

        let functionHandle = this.functionsHandle[functionName];
        try {
            functionHandle(
                msg,
                ctx,
                (err, respMsg) => {
                    if (err != null) {
                        this.logger.error("error when invoking function %s: %s" , functionName, err.toString());
                        return callback(new Error("[UserCodeInvoke]: " + err.toString()));
                    }

                    if (respMsg === "" || respMsg === undefined) {
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
                            return callback(new Error("[UserCodeReturn]: " + error.toString()));
                        }
                    }
                    callback(null, call.request);
                })
        } catch(e) {
            this.logger.error("error when invoking function %s: %s" , functionName, e.toString());
            return callback(new Error("[UserCodeInvoke]: " + e.toString()));
        }
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
