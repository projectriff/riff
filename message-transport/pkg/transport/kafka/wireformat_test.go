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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/projectriff/message-transport/pkg/message"
)

var _ = Describe("Wireformat", func() {

	It("should preserve an empty message", func() {
		msg := message.NewEmptyMessage()
		encodeThenDecode(msg)
	})

	It("should preserve a message with just a payload", func() {
		msg := message.NewMessage([]byte("Hello"), nil)
		encodeThenDecode(msg)
	})

	It("should preserve a message with just headers", func() {
		msg := message.NewMessage(nil, map[string][]string{"key1":{"value1"}, "k2":{"18.0", "other"}})
		encodeThenDecode(msg)
	})

})

func encodeThenDecode(msg message.Message) {
	bytes, err := encodeMessage(msg)
	Expect(err).NotTo(HaveOccurred())

	msg2, err := extractMessage(bytes)
	Expect(err).NotTo(HaveOccurred())

	Expect(msg).To(Equal(msg2))
}
