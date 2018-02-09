gRPC Node.js
============

Create node module wrapping function.proto

`npm install --save @projectriff/function-proto`

```js
const { FunctionInvokerService, FunctionInvokerClient } = require('@projectriff/function-proto');
const grpc = require('grpc');
const address = '127.0.0.1:50051';

// server
const server = new grpc.Server();
server.addService(FunctionInvokerService, {
    call(call) {
        // TODO implement service
    }
});
server.bind(address, grpc.ServerCredentials.createInsecure());
server.start();

// client
const client = new FunctionInvokerClient(address, grpc.credentials.createInsecure());
```

See the official [gRPC for Node.js](https://grpc.io/docs/quickstart/node.html) reference guide.
