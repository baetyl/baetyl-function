// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var function_pb = require('./function_pb.js');

function serialize_baetyl_Message(arg) {
  if (!(arg instanceof function_pb.Message)) {
    throw new Error('Expected argument of type baetyl.Message');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_baetyl_Message(buffer_arg) {
  return function_pb.Message.deserializeBinary(new Uint8Array(buffer_arg));
}


// The function server definition.
var FunctionService = exports.FunctionService = {
  call: {
    path: '/baetyl.Function/Call',
    requestStream: false,
    responseStream: false,
    requestType: function_pb.Message,
    responseType: function_pb.Message,
    requestSerialize: serialize_baetyl_Message,
    requestDeserialize: deserialize_baetyl_Message,
    responseSerialize: serialize_baetyl_Message,
    responseDeserialize: deserialize_baetyl_Message,
  },
};

exports.FunctionClient = grpc.makeGenericClientConstructor(FunctionService);
