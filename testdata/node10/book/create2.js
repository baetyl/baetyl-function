#!/usr/bin/env node

exports.handler = (event, context, callback) => {
    let a = {
        'name': 'xxxx'
    }
    callback(null, {
        "isBase64Encoded": false,
        "statusCode": '200',
        "headers": { "X-Custom-Header": "headerValue" },
        "body": JSON.stringify(a)
    });
}