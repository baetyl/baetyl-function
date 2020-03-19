#!/usr/bin/env node

exports.handler = (event, context, callback) => {
    throw new Error('test custom error')
    let obj = {
        "isBase64Encoded": false,
        "statusCode": 200,
        "headers": {
            "Content-Type": "application/json"
        },
        "body": 's'
    }
    callback(null, JSON.stringify(obj));
};