/*
 * Copyright 2017 the original author or authors.
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

package dispatcher

import "fmt"

type SynchDispatcher interface {
	Dispatch(in *Message) (*Message, error)
}

type Dispatcher interface {

	Input() chan<- Message

	Output() <-chan Message
}

type Headers map[string]interface{}

type Message struct {
	Payload interface{}
	Headers Headers
}

func (msg Message) String() string {
	switch msg.Payload.(type) {
	case []byte:
		return fmt.Sprintf("Message{%v, %v}", string(msg.Payload.([]byte)), msg.Headers)
	default:
		return fmt.Sprintf("Message{%v, %v}", msg.Payload, msg.Headers)
	}
}

func (h Headers) GetOrDefault(key string, value interface{}) interface{} {
	if v, ok := h[key] ; ok {
		return v
	} else {
		return value
	}
}
