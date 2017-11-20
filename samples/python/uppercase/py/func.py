import os,sys
sys.path.insert(0, os.path.abspath('.'))

import grpc
import time
import function_pb2_grpc as function
import fntypes_pb2 as types
from concurrent import futures

class StringFunctionServicer(function.StringFunctionServicer):

    def Call(self, request, context):
        reply = types.Reply()
        reply.body = request.body.upper()
        return reply

server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
function.add_StringFunctionServicer_to_server(StringFunctionServicer(), server)
server.add_insecure_port('%s:%s' % ('[::]', os.environ.get("GRPC_PORT","10382")))

server.start()

while True:
    time.sleep(10)
