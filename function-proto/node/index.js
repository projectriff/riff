/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const grpc = require('grpc');
const path = require('path');

const fn = grpc.load(path.resolve(__dirname, 'function.proto')).function;

function MessageBuilder(headers, payload) {
    this._headers = headers || {};
    this._payload = payload || null;
}
MessageBuilder.prototype = {
    addHeader(name, value) {
        return new MessageBuilder(
            Object.assign({}, this._headers, { [name]: [...(this._headers[name] || []), '' + value] }),
            this._payload
        );
    },
    payload(payload) {
        return new MessageBuilder(this._headers, payload);
    },
    build() {
        return {
            headers: Object.keys(this._headers).reduce((headers, name) => {
                headers[name] = { values: this._headers[name] };
                return headers;
            }, {}),
            payload: Buffer.from(this._payload == null ? [] : this._payload)
        };
    }
};

module.exports = {
    MessageBuilder,
    FunctionInvokerService: fn.MessageFunction.service,
    FunctionInvokerClient: fn.MessageFunction
};
