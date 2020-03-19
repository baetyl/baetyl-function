#!/usr/bin/env node

exports.handler = (event, context, callback) => {
    let obj = {
        "isBase64Encoded": false,
        "statusCode": 200,
        "headers": {
            "Content-Type": "application/json"
        },
        "body": 's',
        'custom': {},
        'custom2': '',
        'custom3': 12,
        'custom4': False,
    };
    callback(null, JSON.stringify(obj));
};