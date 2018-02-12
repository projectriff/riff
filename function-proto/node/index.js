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

function cloneMap(src) {
    const dest = new Map();
    for (const [key, value] of src.entries()) {
        dest.set(key, Array.isArray(value) ? value.slice() : value);
    }
    return dest;
}

function MessageHeaders(headers) {
    if (headers instanceof MessageHeaders) {
        this._names = cloneMap(headers._names);
        this._values = cloneMap(headers._values);
    } else {
        this._names = new Map();
        this._values = new Map();
    }
}
MessageHeaders.fromObject = obj => {
    let headers = new MessageHeaders();
    for (const name of Object.keys(obj)) {
        const values = obj[name].values;
        headers = headers.addHeader(name, ...values);
    }
    return headers;
};
MessageHeaders.prototype = {
    _normalize(name) {
        return name.toLowerCase();
    },
    addHeader(name, ...values) {
        const normalName = this._normalize(name);
        const next = new MessageHeaders(this);
        if (!next._names.has(normalName)) {
            next._names.set(normalName, name);
        }
        if (next._values.has(normalName)) {
            values = this._values.get(normalName).concat(values);
        }
        next._values.set(normalName, values);
        return next;
    },
    replaceHeader(name, ...values) {
        const normalName = this._normalize(name);
        const next = new MessageHeaders(this);
        next._names.set(normalName, name);
        next._values.set(normalName, values);
        return next;
    },
    getValue(name) {
        const normalName = this._normalize(name);
        const values = this._values.get(normalName);
        return values ? values[0] : null;
    },
    getValues(name) {
        const normalName = this._normalize(name);
        return this._values.get(normalName).slice();
    },
    toObject() {
        const output = {};
        for (const normalName of this._names.keys()) {
            output[this._names.get(normalName)] = {
                values: this._values.get(normalName).map(v => '' + v)
            };
        }
        return output;
    }
};

function MessageBuilder(headers, payload) {
    this._headers = headers || new MessageHeaders();
    this._payload = payload || null;
}
MessageBuilder.prototype = {
    addHeader(name, ...value) {
        return new MessageBuilder(
            this._headers.addHeader(name, ...value),
            this._payload
        );
    },
    replaceHeader(name, ...value) {
        return new MessageBuilder(
            this._headers.replaceHeader(name, ...value),
            this._payload
        );
    },
    payload(payload) {
        return new MessageBuilder(this._headers, payload);
    },
    build() {
        return {
            headers: this._headers.toObject(),
            payload: Buffer.from(this._payload == null ? [] : this._payload)
        };
    }
};

module.exports = {
    MessageBuilder,
    MessageHeaders,
    FunctionInvokerService: fn.MessageFunction.service,
    FunctionInvokerClient: fn.MessageFunction
};
