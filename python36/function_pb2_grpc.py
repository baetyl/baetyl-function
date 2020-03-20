# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import function_pb2 as function__pb2


class FunctionStub(object):
  """The function server definition.
  """

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.Call = channel.unary_unary(
        '/faas.Function/Call',
        request_serializer=function__pb2.Message.SerializeToString,
        response_deserializer=function__pb2.Message.FromString,
        )


class FunctionServicer(object):
  """The function server definition.
  """

  def Call(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_FunctionServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'Call': grpc.unary_unary_rpc_method_handler(
          servicer.Call,
          request_deserializer=function__pb2.Message.FromString,
          response_serializer=function__pb2.Message.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'faas.Function', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
