FROM python:3.6-slim
ADD  py/requirements.txt /
RUN  ["pip","install","-r","requirements.txt"]
ADD  proto /proto
# Generate the protobufs
RUN ["python", "-m", "grpc_tools.protoc","-I./proto","--python_out=.", "--grpc_python_out=.","./proto/function.proto"]
ADD py/func.py /
ENTRYPOINT ["python","./func.py"]
