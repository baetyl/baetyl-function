#!/usr/bin/env node

exports.handler = (event, context, callback) => {
    callback(null, {
        "statusCode": 200,
        "headers": {
            "X-Custom-Header": "headers"
        },
        "body": "baetyl"
    });
}