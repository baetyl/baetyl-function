#!/usr/bin/env python3
# -*- coding:utf-8 -*-
"""
python3 runtime
"""

import argparse
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
_CONF_PATH          = "/etc/baetyl/service.yml"
_CODE_PATH          = "/var/lib/baetyl"
_SERVER_ADDRESS     = "0.0.0.0:80"
_HeaderDelim        = "&__header_delim__&"
_HeaderEquals       = "&__header_equals__&"


class mo(function_pb2_grpc.FunctionServicer):
    """
    grpc server module for python3 runtime
    """

    def Load(self, conf):
        """
        load config and init module
        """
        self.config = {}
        if os.path.exists(conf):
            self.config = yaml.load(
                open(conf, 'r').read(), Loader=yaml.FullLoader)

        if 'SERVICE_NAME' in os.environ:
            self.config['name'] = os.environ['SERVICE_NAME']
        else:
            raise Exception('config invalid, missing name in os.environ')

        self.config['server'] = {
            'address': _SERVER_ADDRESS
        }

        self.log = get_logger(self.config)
        self.functions = get_functions(_CODE_PATH)
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

    def Invoke(self, request, context):
        """
        call request
        """
        if request.Type == 'HTTP':
            return process_http(request, context)
        else:
            raise Exception("Type of MessageRequest doesn't support")

    def process_http(self, request, context):
        event = parse_http_params(request)
        ctx = {}
        ctx['invokeid'] = request.Metadata['invokeId']
        ctx['functionName'] = request.Name

        method = request.Method
        if method == "":
            method = self.functions[0]
        if method not in self.functions:
            return populate_http_response(404, "no router")

        try:
            response = self.functions[method](event, ctx)
        except BaseException as err:
            self.log.error(err, exc_info=True)
            return populate_http_response(500, err)
        
        if !(isinstance(response, dict) or isinstance(response, str)):
            return populate_http_response(502, "function response error")

        if isinstance(response, str):
            try:
                response =json.loads(response)
            except BaseException:
                return populate_http_response(502, "function response error")

        message = {
            'Metadata': {}
        }

        if 'statusCode' not in response or not isinstance(response['statusCode'], int):
            return populate_http_response(502, "function response error")
        else:
            message['Metadata']['statusCode'] = response['statusCode']
        
        if 'headers' in response:
            try:
                headers = json.loads(response['headers'])
            except BaseException:
                return populate_http_response(502, "function response error")
            items = []
            for k, v in headers.items():
                if isinstance(v, str):
                    return populate_http_response(502, "function response error")
                items.append(k + _HeaderEquals + v)
            message['Metadata']['headers'] = _HeaderDelim.join(items)

        if 'isBase64Encoded' in response:
            if isinstance(response['isBase64Encoded'], bool):
                message['Metadata']['isBase64Encoded'] = response['isBase64Encoded']
            else:
                return populate_http_response(502, "function response error")

        if 'body' in response:
            if isinstance(response['body'], str):
                message['Payload'] = bytes(response['body'], encoding = "utf8")
            else:
                return populate_http_response(502, "function response error")

        return response


def get_functions(code_path):
    functions_handler = {}
    for root, dirs, files in os.walk(code_path):
        for name in files:
            if os.path.splitext(name)[-1] == ".yml":
                load_functions(root, name, functions_handler)
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
        sys.path.append()
        module_handler = fc['handler'].split('.')
        handler_name = module_handler.pop()
        module = importlib.import_module(
            os.path.join(root, fc['codedir'], module_handler))
        functions_handler[fc['name']] = getattr(module, handler_name)


def get_grpc_server(c):
    """
    get grpc server
    """
    # TODO: to test
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
    logger = logging.getLogger(c['name'])
    if 'logger' not in c:
        return logger
    if 'path' not in c['logger']:
        return logger

    filename = os.path.abspath(c['logger']['path'])
    os.makedirs(os.path.dirname(filename), exist_ok=True)

    level = logging.INFO
    if 'level' in c['logger']:
        if c['logger']['level'] == 'debug':
            level = logging.DEBUG
        elif c['logger']['level'] == 'warn':
            level = logging.WARNING
        elif c['logger']['level'] == 'error':
            level = logging.ERROR

    interval = 15
    if 'age' in c['logger'] and 'max' in c['logger']['age']:
        interval = c['logger']['age']['max']

    backupCount = 15
    if 'backup' in c['logger'] and 'max' in c['logger']['backup']:
        backupCount = c['logger']['backup']['max']

    logger.setLevel(level)

    # create a file handler
    handler = logging.handlers.TimedRotatingFileHandler(
        filename=filename,
        when='h',
        interval=interval,
        backupCount=backupCount)
    handler.setLevel(level)

    formatter = logging.Formatter(
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)

    logger.addHandler(handler)
    return logger


def parse_http_params(request):
    event = {}
    event['path'] = request.Metadata['path']
    event['resource'] = event['path']
    event['httpMethod'] = request.Metadata['httpMethod']
    event['pathParameters'] = {}
    event['body'] = request.Payload.decode()
    event['isBase64Encoded'] = request.Metadata['isBase64Encoded']
    event['queryStringParameters'] = parse.parse_qs(
        event['queryStringParameters'])
    event['headers'] = request.Metadata['headers']
    headers = {}
    for header in event['headers'].split(_HeaderDelim):
        kv = header.split(_HeaderEquals)
        headers[kv[0]] = kv[1]
    event['headers'] = headers
    event['requestContext'] = {
        "stage": "",
		"requestId": request.Metadata['invokeId'],
		"resourcePath": event['resource'],
		"httpMethod": event['httpMethod'],
		"apiId": "",
        "sourceIp": "",
    }
    return event

def populate_http_response(code, err, body=None):
    if code == 502:
        body = {
            "Code": "BadGatewayException",
            "Cause": err,
            "Message": "Bad Gateway",
            "Status": 502,
            "Type": "Server"
        }
    elif code == 404:
        body = {
            "Code": "NotFountException",
            "Cause": err,
            "Message": "Not Found",
            "Status": 404,
            "Type": "User"
        }
    elif code == 500:
        body = {
            "errorMessage": err
        }
        
    response =  {
        'Metadata': {
            'statusCode': code
        },
        'Payload': json.dumps(body)
    }
    return response


if __name__ == '__main__':
    m = mo()
    m.Load(_CONF_PATH)
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
