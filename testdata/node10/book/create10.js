#!/usr/bin/env node

exports.handler = (event, context, callback) => {
    let obj = {
        "name": "baetyl",
        "project": "github"
    }
    callback(null, {
        "isBase64Encoded": false,
        "statusCode": 200,
        "headers": {
            "X-Custom-Header": "headers"
        },
        "body": JSON.stringify(obj)
    });
}