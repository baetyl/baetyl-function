#!/usr/bin/env python3
# -*- coding:utf-8 -*-
"""
python3 runtime
"""

import importlib
import os
import sys
import time
import grpc
import yaml
import json
import signal
from concurrent import futures
import function_pb2
import function_pb2_grpc
import logging
import logging.handlers
from urllib import parse

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


class mo(function_pb2_grpc.FunctionServicer):
    """
    grpc server module for python3 runtime
    """

    def __init__(self):
        self.name = 'baetyl-function'
        self.conf_path = '/etc/baetyl/service.yml'
        self.code_path = '/var/lib/baetyl/code'
        self.server_address = "0.0.0.0:80"

    def Load(self):
        """
        load config and init module
        """
        if 'SERVICE_NAME' in os.environ:
            self.name = os.environ['SERVICE_NAME']

        if 'SERVICE_CONF' in os.environ:
            self.conf_path = os.environ['SERVICE_CONF']

        if 'SERVICE_CODE' in os.environ:
            self.code_path = os.environ['SERVICE_CODE']

        if 'SERVICE_ADDRESS' in os.environ:
            self.server_address = os.environ['SERVICE_ADDRESS']

        self.config = {
            'server': {
                'address': self.server_address
            }
        }

        if os.path.exists(self.conf_path):
            self.config = yaml.load(
                open(self.config, 'r').read(), Loader=yaml.FullLoader)

        self.log = get_logger(self)
        self.functions = get_functions(self.code_path)
        self.server = get_grpc_server(self.config['server'])
        function_pb2_grpc.add_FunctionServicer_to_server(self, self.server)

    def Start(self):
        """
        start module
        """
        self.log.info("service starting")
        self.server.start()

    def Close(self):
        """
        close module
        """
        grace = None
        if 'timeout' in self.config['server']:
            grace = self.config['server']['timeout'] / 1e9
        self.server.stop(grace)
        self.log.info("service closed")

    def Call(self, request, context):
        """
        call request
        """
        method = request.Metadata['method']
        if method == "":
            method = list(self.functions.keys())[0]

        if method not in self.functions:
            self.log.error("method not found: %s", method)
            raise Exception('method not found')

        ctx = {}
        for k in request.Metadata.keys():
            ctx[k] = request.Metadata[k]
        ctx["id"] = request.ID

        msg = b''
        if request.Payload:
            try:
                msg = json.loads(request.Payload)
            except BaseException:
                msg = request.Payload  # raw data, not json format

        try:
            msg = self.functions[method](msg, ctx)
        except BaseException as err:
            self.log.error("error when invoke method %s: %s", method, err)
            raise Exception("[UserCodeInvoke] ", err)

        if msg is None:
            request.Payload = b''
        elif isinstance(msg, bytes):
            request.Payload = msg
        else:
            try:
                request.Payload = json.dumps(msg).encode('utf-8')
            except BaseException as err:
                self.log.error(err, exc_info=True)
                raise Exception("[UserCodeReturn] ", err)
        return request


def get_functions(code_path):
    functions_handler = {}
    for root, dirs, files in os.walk(code_path):
        sys.path.append(code_path)
        for name in files:
            if os.path.splitext(name)[-1] == ".yml":
                load_functions(root, name, functions_handler)
        break
    return functions_handler


def load_functions(root, name, functions_handler):
    config = yaml.load(open(os.path.join(root, name),
                            'r').read(), Loader=yaml.FullLoader)
    if 'functions' not in config:
        raise Exception('config invalid, missing functions')

    for fc in config['functions']:
        if 'name' not in fc or 'handler' not in fc or 'codedir' not in fc:
            raise Exception(
                'config invalid, missing function name, handler or codedir')
        module_spec = fc['handler'].split('.')
        handler_name = module_spec.pop()
        module_name = module_spec.pop()
        module = importlib.import_module(
            os.path.join(fc['codedir'], module_name).replace('./', '').replace('/', '.'))
        functions_handler[fc['name']] = getattr(module, handler_name)


def get_grpc_server(c):
    """
    get grpc server
    """
    max_workers = None
    max_concurrent = None
    max_message_length = 4 * 1024 * 1024
    if 'workers' in c:
        if 'max' in c['workers']:
            max_workers = c['workers']['max']
    if 'concurrent' in c:
        if 'max' in c['concurrent']:
            max_concurrent = c['concurrent']['max']
    if 'message' in c:
        if 'length' in c['message']:
            if 'max' in c['message']['length']:
                max_message_length = c['message']['length']['max']

    ssl_ca = None
    ssl_key = None
    ssl_cert = None
    if 'ca' in c:
        with open(c['ca'], 'rb') as f:
            ssl_ca = f.read()
    if 'key' in c:
        with open(c['key'], 'rb') as f:
            ssl_key = f.read()
    if 'cert' in c:
        with open(c['cert'], 'rb') as f:
            ssl_cert = f.read()

    s = grpc.server(thread_pool=futures.ThreadPoolExecutor(max_workers=max_workers),
                    options=[('grpc.max_send_message_length', max_message_length),
                             ('grpc.max_receive_message_length', max_message_length)],
                    maximum_concurrent_rpcs=max_concurrent)
    if ssl_key is not None and ssl_cert is not None:
        credentials = grpc.ssl_server_credentials(
            ((ssl_key, ssl_cert),), ssl_ca, ssl_ca is not None)
        s.add_secure_port(c['address'], credentials)
    else:
        s.add_insecure_port(c['address'])
    return s


def get_logger(c):
    """
    get logger
    """
    logging.basicConfig(level=logging.INFO,
                        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    logger = logging.getLogger(c.name)
    if 'logger' not in c.config:
        return logger
    if 'path' not in c.config['logger']:
        return logger

    filename = os.path.abspath(c.config['logger']['path'])
    os.makedirs(os.path.dirname(filename), exist_ok=True)

    if 'level' in c.config['logger']:
        if c.config['logger']['level'] == 'debug':
            level = logging.DEBUG
        elif c.config['logger']['level'] == 'warn':
            level = logging.WARNING
        elif c.config['logger']['level'] == 'error':
            level = logging.ERROR

    interval = 15
    if 'age' in c.config['logger'] and 'max' in c.config['logger']['age']:
        interval = c.config['logger']['age']['max']

    backup_count = 15
    if 'backup' in c.config['logger'] and 'max' in c.config['logger']['backup']:
        backup_count = c.config['logger']['backup']['max']

    logger.setLevel(level)

    # create a file handler
    handler = logging.handlers.TimedRotatingFileHandler(
        filename=filename,
        when='h',
        interval=interval,
        backupCount=backup_count)
    handler.setLevel(level)

    formatter = logging.Formatter(
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)

    logger.addHandler(handler)
    return logger


if __name__ == '__main__':
    m = mo()
    m.Load()
    m.Start()

    def exit(signum, frame):
        sys.exit(0)

    signal.signal(signal.SIGINT, exit)
    signal.signal(signal.SIGTERM, exit)

    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except BaseException as err:
        m.log.debug(err)
    finally:
        m.Close()
