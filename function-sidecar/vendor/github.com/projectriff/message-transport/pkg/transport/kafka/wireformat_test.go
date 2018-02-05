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

package kafka

import (
	"testing"
	"reflect"
	"github.com/projectriff/message-transport/pkg/message"
)

func TestEmptyMessage(t *testing.T) {
	msg := message.NewEmptyMessage()
	encodeThenDecode(msg, t)
}

func TestMessageWithOnlyPayload(t *testing.T) {
	msg := message.NewMessage([]byte("Hello"), nil)
	encodeThenDecode(msg, t)
}

func TestMessageWithOnlyHeaders(t *testing.T) {
	msg := message.NewMessage(nil, map[string][]string{"key1":{"value1"}, "k2":{"18.0", "other"}})
	encodeThenDecode(msg, t)
}

func encodeThenDecode(msg message.Message, t *testing.T) {
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