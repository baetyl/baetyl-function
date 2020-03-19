#!/usr/bin/env python
# -*- coding:utf-8 -*-
"""
function to say hi in python
"""

import json

def handler(event, context):
    resp = {
        "isBase64Encoded": False,
        "statusCode": 200,
        "headers": {
	        "Content-Type": "application/json"
        },
        "body": 's',
        'custom': {},
        'custom2': '',
        'custom3': 12,
        'custom4': false,
    }
    return resp
