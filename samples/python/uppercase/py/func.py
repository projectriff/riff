import os,sys
sys.path.insert(0, os.path.abspath('.'))

import grpc
import time
import function_pb2_grpc as function
import function_pb2 as message
from concurrent import futures

'''
This method’s semantics are a combination of those of the request-streaming method and the response-streaming method.
It is passed an iterator of request values and is itself an iterator of response values.
'''
class MessageFunctionServicer(function.MessageFunctionServicer):

    def Call(self, request_iterator, context):
        for request in request_iterator:
            reply = message.Message()
            reply.payload = request.payload.upper()
            reply.headers['correlationId'].values[:] = request.headers['correlationId'].values[:]
            yield reply

server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
function.add_MessageFunctionServicer_to_server(MessageFunctionServicer(), server)
server.add_insecure_port('%s:%s' % ('[::]', os.environ.get("GRPC_PORT","10382")))

server.start()

while True:
    time.sleep(10)
