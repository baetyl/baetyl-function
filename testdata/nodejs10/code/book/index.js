#!/usr/bin/env node

const hasAttr = (obj, attr) => {
    if (obj instanceof Object && !(obj instanceof Array)) {
        if (obj[attr] != undefined) {
            return true;
        }
    }
    return false;
};

const passParameters = (event, context) => {
    if (hasAttr(context, 'messageQOS')) {
        event['messageQOS'] = context['messageQOS'];
    }
    if (hasAttr(context, 'messageTopic')) {
        event['messageTopic'] = context['messageTopic'];
    }
    if (hasAttr(context, 'messageTimestamp')) {
        event['messageTimestamp'] = context['messageTimestamp'];
    }
    if (hasAttr(context, 'functionName')) {
        event['functionName'] = context['functionName'];
    }
    if (hasAttr(context, 'invokeid')) {
        event['invokeid'] = context['invokeid'];
    }
};

exports.handler = (event, context, callback) => {
    // support Buffer & json object
    if (Buffer.isBuffer(event)) {
        const message = event.toString();
        event = {}
        event["bytes"] = message;
    }
    else if(hasAttr(event, 'err')) {
        return callback(new TypeError(event['err']))
    }

    passParameters(event, context);
    event['Say'] = 'Hello Baetyl';
    callback(null, event);
};