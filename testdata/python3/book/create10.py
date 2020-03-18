#!/usr/bin/env python
# -*- coding:utf-8 -*-
"""
function to say hi in python
"""

import json

def handler(event, context):
    obj = {
        "name": "baetyl",
        "project": "github"
    }
    resp = {
        "isBase64Encoded": False,
        "statusCode": 200,
        "headers": {
            "X-Custom-Header": "headers"
        },
        "body": json.dumps(obj)
    }
    return resp