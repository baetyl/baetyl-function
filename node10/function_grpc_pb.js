// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var function_pb = require('./function_pb.js');

function serialize_baetyl_MessageRequest(arg) {
  if (!(arg instanceof function_pb.MessageRequest)) {
    throw new Error('Expected argument of type baetyl.MessageRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_baetyl_MessageRequest(buffer_arg) {
  return function_pb.MessageRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// The function server definition.
var FunctionService = exports.FunctionService = {
  call: {
    path: '/baetyl.Function/Call',
    requestStream: false,
    responseStream: false,
    requestType: function_pb.MessageRequest,
    responseType: function_pb.MessageRequest,
    requestSerialize: serialize_baetyl_MessageRequest,
    requestDeserialize: deserialize_baetyl_MessageRequest,
    responseSerialize: serialize_baetyl_MessageRequest,
    responseDeserialize: deserialize_baetyl_MessageRequest,
  },
};

exports.FunctionClient = grpc.makeGenericClientConstructor(FunctionService);
