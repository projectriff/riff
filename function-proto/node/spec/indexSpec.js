describe('function-proto', () => {
    const { FunctionInvokerService, FunctionInvokerClient, MessageBuilder, MessageHeaders } = require('..');
    const grpc = require('grpc');
    const port = 50051;

    describe('MessageHeaders', () => {
        it('parses a plain JS object', () => {
            const obj = {
                'Content-Type': {
                    values: ['application/json']
                }
            }
            const headers = MessageHeaders.fromObject(obj);
            expect(headers.getValue('content-type')).toBe('application/json')
        });

        it('creates a plain JS object', () => {
            const headers = new MessageHeaders()
                .addHeader('Content-Type', 'application/json');
            expect(headers.toObject()).toEqual({
                'Content-Type': {
                    values: ['application/json']
                }
            });
        });

        it('creates a plain JS object', () => {
            const headers = new MessageHeaders()
                .addHeader('Content-Type', 'application/json');
            expect(headers.toObject()).toEqual({
                'Content-Type': {
                    values: ['application/json']
                }
            });
        });

        it('allows for multiple values', () => {
            const headers = new MessageHeaders()
                .addHeader('Accept', '*/*;q=0.1', 'text/plain;q=0.9');
            expect(headers.getValues('accept')).toEqual(['*/*;q=0.1', 'text/plain;q=0.9']);
        });

        it('returns the first value for a header', () => {
            const headers = new MessageHeaders()
                .addHeader('Accept', '*/*;q=0.1', 'text/plain;q=0.9');
            expect(headers.getValue('accept')).toEqual('*/*;q=0.1');
        });

        it('conflates header names to be case insensite', () => {
            const headers = new MessageHeaders()
                .addHeader('Accept', '*/*;q=0.1')
                .addHeader('accept', 'text/plain;q=0.9');
            expect(headers.toObject()).toEqual({
                'Accept': {
                    values: ['*/*;q=0.1', 'text/plain;q=0.9']
                }
            });
        });

        it('converts header values to strings', () => {
            const headers = new MessageHeaders()
                .addHeader('correlationId', 1234);
            expect(headers.toObject()).toEqual({
                'correlationId': {
                    values: ['1234']
                }
            });
        });

        it('handles multiple headers', () => {
            const headers = new MessageHeaders()
                .addHeader('Accept', '*/*;q=0.1')
                .addHeader('accept', 'text/plain;q=0.9')
                .addHeader('Content-Type', 'application/json');
            expect(headers.toObject()).toEqual({
                'Accept': {
                    values: ['*/*;q=0.1', 'text/plain;q=0.9']
                },
                'Content-Type': {
                    values: ['application/json']
                }
            });
        });

        it('is immutable', () => {
            const empty = new MessageHeaders();
            const one = empty.addHeader('correlationId', '1234');
            expect(one).not.toBe(empty);
            expect(empty.toObject()).toEqual({});
            expect(one.toObject()).toEqual({
                correlationId: {
                    values: ['1234']
                }
            });
        });
    });

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
