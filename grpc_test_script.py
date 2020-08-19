import grpc

from service_pb2_grpc import HelloWorldServiceStub
from service_pb2 import HelloRequest

channel = grpc.insecure_channel('localhost:50051')
hello_world_stub = HelloWorldServiceStub(channel)

while True:
    name = input("Enter your name: ")
    request = HelloRequest()
    request.name = name
    response = hello_world_stub.Hello(request)
    print(response.message)
