// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var function_pb = require('./function_pb.js');

function serialize_faas_Message(arg) {
  if (!(arg instanceof function_pb.Message)) {
    throw new Error('Expected argument of type faas.Message');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_faas_Message(buffer_arg) {
  return function_pb.Message.deserializeBinary(new Uint8Array(buffer_arg));
}


// The function server definition.
var FunctionService = exports.FunctionService = {
  call: {
    path: '/faas.Function/Call',
    requestStream: false,
    responseStream: false,
    requestType: function_pb.Message,
    responseType: function_pb.Message,
    requestSerialize: serialize_faas_Message,
    requestDeserialize: deserialize_faas_Message,
    responseSerialize: serialize_faas_Message,
    responseDeserialize: deserialize_faas_Message,
  },
};

exports.FunctionClient = grpc.makeGenericClientConstructor(FunctionService);
