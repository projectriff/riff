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

package grpc

import "github.com/projectriff/riff/message-transport/pkg/message"

// Type size is a WindowingStrategy based on the number of input messages seen: the stream is closed after N messages.
type size int

// factory
func sizeFactoryFactory(wanted int) WindowingStrategyFactory {
	return func (m message.Message) WindowingStrategy {
		s := size(wanted)
		return &s
	}
}

func (s *size) ShouldClose(in message.Message, err error) bool {
	*s--
	return *s == 0
}

func (s *size) AsyncClosingChannel() <-chan struct{} {
	return nil
}