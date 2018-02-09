describe('function-proto', () => {
    const { FunctionInvokerService, FunctionInvokerClient, MessageBuilder } = require('..');
    const grpc = require('grpc');
    const port = 50051;

    it('builds a message', () => {
        const theMessage = new MessageBuilder()
            .addHeader('Header-Name', 'headerValue 1')
            .addHeader('Header-Name', 'headerValue 2')
            .payload('will be replaced')
            .payload('riff')
            .build();

        expect(theMessage).toEqual({
            headers: {
                'Header-Name': {
                    values: [
                        'headerValue 1',
                        'headerValue 2'
                    ]
                }
            },
            payload: Buffer.from('riff')
        });
    });

    it('builds an empty message', () => {
        const theMessage = new MessageBuilder().build();

        expect(theMessage).toEqual({
            headers: {},
            payload: Buffer.from([])
        });
    });

    it('generates a grpc client and server', done => {
        const theMessage = new MessageBuilder()
            .addHeader('Header-Name', 'headerValue')
            .payload('riff')
            .build();

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
