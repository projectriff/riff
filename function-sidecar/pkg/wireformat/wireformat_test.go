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

package wireformat

import (
	"testing"
	"reflect"
	"github.com/projectriff/function-sidecar/pkg/dispatcher"
)

func TestEmptyMessage(t *testing.T) {
	msg := dispatcher.Message{}
	encodeThenDecode(msg, t)
}

func TestMessageWithOnlyPayload(t *testing.T) {
	msg := dispatcher.Message{Payload: []byte("Hello")}
	encodeThenDecode(msg, t)
}

func TestMessageWithOnlyHeaders(t *testing.T) {
	msg := dispatcher.Message{Headers: map[string]interface{}{"key1":"value1", "k2":18.0}}
	encodeThenDecode(msg, t)
}

func encodeThenDecode(msg dispatcher.Message, t *testing.T) {
	bytes, err := encodeMessage(msg)
	if err != nil {
		t.Fatal(err)
	}
	msg2, err := extractMessage(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(msg, msg2) {
		t.Fatal("Expected identical messages: ", msg, msg2)
	}
}