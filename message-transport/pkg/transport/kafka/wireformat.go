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
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/Shopify/sarama"
	"github.com/projectriff/message-transport/pkg/message"
)

// This file serializes/deserializes a message.Message on a Kafka topic.
// Currently uses a custom encoding scheme for headers, until Kafka 0.11 headers are supported by go client lib

func fromKafka(kafka *sarama.ConsumerMessage) (message.Message, error) {
	return extractMessage(kafka.Value)
}

func toKafka(message message.Message) (*sarama.ProducerMessage, error) {
	bytesOut, err := encodeMessage(message)
	if err != nil {
		return nil, err
	}
	return &sarama.ProducerMessage{Value: sarama.ByteEncoder(bytesOut)}, nil
}

func extractMessage(bytes []byte) (message.Message, error) {
	offset := uint32(0)
	if bytes[offset] != 0xff {
		return nil, errors.New("expected 0xff as the leading byte")
	}
	offset++

	headerCount := bytes[offset]
	offset++

	headers := make(map[string][]string, headerCount)
	for i := byte(0); i < headerCount; i = i + 1 {
		len := uint32(bytes[offset])
		offset++

		name := string(bytes[offset : offset+len])
		offset += len

		len = binary.BigEndian.Uint32(bytes[offset : offset+4])
		offset += 4
		var value []string
		err := json.Unmarshal(bytes[offset:offset+len], &value)
		if err != nil {
			return nil, err
		}
		headers[name] = value
		offset += len
	}
	return message.NewMessage(bytes[offset:], headers), nil
}

func encodeMessage(message message.Message) ([]byte, error) {
	length := 0
	length++ // initial 0xff
	length++ // no of headers

	headerValues := make(map[string][]byte, len(message.Headers()))
	for k, v := range message.Headers() {
		length += 1 // 1 byte to encode len(k)
		length += len(k)
		var err error
		headerValues[k], err = json.Marshal(v) // will marshal as json array
		if err != nil {
			return nil, err
		}
		length += 4 // 4bytes to encode len(hv[i])
		length += len(headerValues[k])
	}

	length += len(message.Payload())

	result := make([]byte, length)
	offset := 0

	result[offset] = 0xff
	offset++

	result[offset] = byte(len(message.Headers()))
	offset++

	for k, _ := range message.Headers() {
		l := len(k)
		result[offset] = byte(l)
		offset++

		copy(result[offset:offset+l], []byte(k))
		offset += l

		binary.BigEndian.PutUint32(result[offset:offset+4], uint32(len(headerValues[k])))
		offset += 4
		copy(result[offset:], headerValues[k])
		offset += len(headerValues[k])
	}
	copy(result[offset:], message.Payload())
	return result, nil
}
