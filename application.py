import asyncio
import grpc
import logging

from aiohttp import web
from grpc.experimental.aio import init_grpc_aio
from service_pb2 import HelloResponse
from service_pb2_grpc import add_HelloWorldServiceServicer_to_server
from service_pb2_grpc import HelloWorldServiceServicer


class HelloWorldView(web.View):
    async def get(self) -> web.Response:
        return web.Response(text="Hello World!")


class Application(web.Application):
    def __init__(self):
        super().__init__()

        self.grpc_task = None
        self.grpc_server = GrpcServer()

        self.add_routes()
        self.on_startup.append(self.__on_startup())
        self.on_shutdown.append(self.__on_shutdown())

    def __on_startup(self):
        async def _on_startup(app):
            self._grpc_task = asyncio.ensure_future(app.grpc_server.start())

        return _on_startup

    def __on_shutdown(self):
        async def _on_shutdown(app):
            await app.grpc_server.stop()
            app.grpc_task.cancel()
            await app.grpc_task

        return _on_shutdown

    def add_routes(self):
        self.router.add_view('/helloworld', HelloWorldView)

    def run(self):
        return web.run_app(self, port=8000)


class HelloServicer(HelloWorldServiceServicer):
    def Hello(self, request, context):
        response = HelloResponse()
        response.message = "Hello {}!".format(request.name)
        return response


class GrpcServer:
    def __init__(self):
        init_grpc_aio()

        self.server = grpc.experimental.aio.server()
        self.servicer = HelloServicer()

        add_HelloWorldServiceServicer_to_server(self.servicer, self.server)
        self.server.add_insecure_port("[::]:50051")

    async def start(self):
        await self.server.start()
        await self.server.wait_for_termination()

    async def stop(self):
        await self.server.stop(1)


application = Application()

if __name__ == '__main__':
    application.run()
