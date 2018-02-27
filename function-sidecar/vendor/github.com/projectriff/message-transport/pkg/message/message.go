/*
 * Copyright 2017-Present the original author or authors.
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

// Package message defines an abstract message type with a simple concrete implementation.
package message

import "fmt"

type Message interface {
	Payload() []byte
	Headers() Headers
}

type message struct {
	payload []byte
	headers Headers
}

type Headers map[string][]string

func (msg *message) String() string {
	return fmt.Sprintf("Message{%v, %v}", string(msg.payload), msg.headers)
}

func (msg *message) Payload() []byte {
	return (*msg).payload
}

func (msg *message) Headers() Headers {
	return (*msg).headers
}

func (h Headers) GetOrDefault(key string, value string) string {
	if v, ok := h[key]; ok {
		if len(v) == 0 {
			return value
		} else {
			return v[0]
		}
	} else {
		return value
	}
}

func NewEmptyMessage() Message {
	return NewMessage([]byte{}, map[string][]string{})
}

func NewMessage(payload []byte, headers Headers) Message {
	if payload == nil {
		payload = make([]byte, 0)
	}
	if headers == nil {
		headers = make(map[string][]string, 0)
	}
	return &message{payload: payload, headers: headers}
}
