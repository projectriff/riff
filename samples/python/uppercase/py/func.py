import os,sys
sys.path.insert(0, os.path.abspath('.'))

import grpc
import time
import function_pb2_grpc as function
import fntypes_pb2 as types
from concurrent import futures

'''
This method’s semantics are a combination of those of the request-streaming method and the response-streaming method. 
It is passed an iterator of request values and is itself an iterator of response values.
'''
class StringFunctionServicer(function.StringFunctionServicer):

    def Call(self, request_iterator, context):
        for request in request_iterator:
            reply = types.Reply()
            reply.body = request.body.upper()
            yield reply

server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
function.add_StringFunctionServicer_to_server(StringFunctionServicer(), server)
server.add_insecure_port('%s:%s' % ('[::]', os.environ.get("GRPC_PORT","10382")))

server.start()

while True:
    time.sleep(10)
