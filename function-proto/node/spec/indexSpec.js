describe('function-proto', () => {
    const { FunctionInvokerService, FunctionInvokerClient } = require('..');
    const grpc = require('grpc');
    const port = 50051;

    it('generates a grpc client and server', done => {
        const theMessage = {
            headers: {
                'Header-Name': {
                    values: ['headerValue']
                }
            },
            payload: Buffer.from('riff')
        };

        const server = new grpc.Server();
        server.addService(FunctionInvokerService, {
            call(call) {
                call.on('data', message => {
                    expect(message).toEqual(theMessage);
                    call.write(message);
                });
                call.on('end', () => {
                    call.end();
                });
            }
        });
        server.bind(`127.0.0.1:${port}`, grpc.ServerCredentials.createInsecure());
        server.start();

        // client
        const client = new FunctionInvokerClient(`127.0.0.1:${port}`, grpc.credentials.createInsecure());
        const call = client.call();
        call.on('data', message => {
            expect(message).toEqual(theMessage);
            call.end();
        });
        call.on('end', () => {
            server.tryShutdown(done);
        });
        call.write(theMessage);
    });
});
