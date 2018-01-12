FROM python:3.6-slim as build_proto
ADD  py/grpcio-tools.txt /
RUN  ["pip","install","-r","grpcio-tools.txt"]
ADD  proto /proto
# Generate the protobufs
RUN ["mkdir", "./grpc"]
RUN ["python", "-m", "grpc_tools.protoc","-I./proto","--python_out=./grpc", "--grpc_python_out=./grpc","./proto/function.proto"]

FROM python:3.6-slim
ADD py/func.py /
ADD py/requirements.txt /
RUN  ["pip","install","-r","requirements.txt"]
COPY --from=build_proto ./grpc .
CMD ["./app"]  
ENTRYPOINT ["python","./func.py"]
