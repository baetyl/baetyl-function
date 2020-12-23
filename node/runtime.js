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

const getFunctions = s => {
    let functionsHandle = {};
    if (!hasAttr(s.config, 'functions')) {
        return functionsHandle;
    }

    let cwd = process.cwd();
    s.config.functions.forEach(function (ele) {
        if (!ele.name || !ele.handler) {
            throw new Error('config invalid, missing function name, handler');
        }

        let codedir = !ele.codedir ? '' : ele.codedir;
        const moduleHandler = ele.handler.split('.');
        const handlerName = moduleHandler[1];
        const moduleName = require(path.join(cwd, s.codePath, path.join("/", codedir), moduleHandler[0]));
        functionsHandle[ele.name] = moduleName[handlerName];
    });
    return functionsHandle;
};

const getGrpcServer = s => {
    let config = {
        'address': s.serverAddress,
        'ca': s.cert['ca'],
        'key': s.cert['key'],
        'cert': s.cert['cert'],
    }

    if (hasAttr(s.config, 'server')) {
        config = s.config['server']
    }

    let maxMessageSize = 4 * 1024 * 1024;
    if (hasAttr(config, 'message')
        && hasAttr(config['message'], 'length')
        && hasAttr(config['message']['length'], 'max')) {
        maxMessageSize = config['message']['length']['max'];
    }
    let server = new grpc.Server({
        'grpc.max_send_message_length': maxMessageSize,
        'grpc.max_receive_message_length': maxMessageSize
    });

    let credentials = undefined;

    if (hasAttr(config, 'ca')
        && hasAttr(config, 'key')
        && hasAttr(config, 'cert')) {

        credentials = grpc.ServerCredentials.createSsl(
            fs.readFileSync(config['ca']), [{
                cert_chain: fs.readFileSync(config['cert']),
                private_key: fs.readFileSync(config['key'])
            }], true);
    } else {
        credentials = grpc.ServerCredentials.createInsecure();
    }

    server.bind(config['address'], credentials);
    return server;
};

class NodeRuntimeModule {
    constructor() {
        this.name = 'baetyl-node10';
        this.confPath = 'etc/baetyl/conf.yml';
        this.codePath = 'var/lib/baetyl/code';
        this.serverAddress = "0.0.0.0:80";
        this.cert = {
            'ca': 'var/lib/baetyl/system/certs/ca.pem',
            'key': 'var/lib/baetyl/system/certs/key.pem',
            'cert': 'var/lib/baetyl/system/certs/crt.pem',
        };
    }

    Load() {
        if (!(fs.existsSync(this.cert['ca'])
            && fs.existsSync(this.cert['key'])
            && fs.existsSync(this.cert['cert']))) {
            throw new Error("system certificate is not found");
        }

        if (hasAttr(process.env, 'BAETYL_SERVICE_NAME')) {
            this.name = process.env['BAETYL_SERVICE_NAME']
        } 

        if (hasAttr(process.env, 'BAETYL_CONF_FILE')) {
            this.confPath = process.env['BAETYL_CONF_FILE']
        }

        if (hasAttr(process.env, 'BAETYL_CODE_PATH')) {
            this.codePath = process.env['BAETYL_CODE_PATH']
        }

        if (!hasAttr(process.env, 'BAETYL_RUN_MODE')) {
            throw new Error("BAETYL_RUN_MODE env is not found");
        }

        if (process.env['BAETYL_RUN_MODE'] === 'native') {
            this.serverAddress = "127.0.0.1:" + process.env['BAETYL_SERVICE_DYNAMIC_PORT']
        }

        this.config = {}

        if (fs.existsSync(this.confPath)) {
            this.config = yaml.parse(fs.readFileSync(this.confPath).toString());
        }
        
        this.logger = getLogger(this);
        this.functionsHandle = getFunctions(this);
        this.server = getGrpcServer(this);

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
        if (!functionName) {
            if (Object.keys(this.functionsHandle).length < 1) {
                this.logger.error("no functions exist");
                return callback(new Error("no functions exist"));
            }
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
        try {
            const payloadString = Buffer.from(Payload).toString();
            msg = JSON.parse(payloadString);
        } catch (error) {
            msg = Buffer.from(Payload); // raw data, not json format
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
