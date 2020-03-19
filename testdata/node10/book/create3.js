#!/usr/bin/env node

exports.handler = (event, context, callback) => {
    callback(null, {
        "isBase64Encoded": false,
        "statusCode": 200,
        "headers": 'headers',
        "body": "computer-computer"
    });
}