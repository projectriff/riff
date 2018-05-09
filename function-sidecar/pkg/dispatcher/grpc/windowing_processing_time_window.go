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

import (
	"github.com/projectriff/riff/message-transport/pkg/message"
	"time"
)

// Type ptimeWindow is a WindowingStrategy that is based on elapsed wall clock time.
type ptimeWindow struct {
	end  time.Time
	done chan struct{}
}

func ptimeFactoryFactory(d time.Duration) WindowingStrategyFactory {
	return func(first message.Message) WindowingStrategy {
		c := make(chan struct{})
		go func() {
			<- time.After(d)
			close(c)
		}()
		return &ptimeWindow{end: time.Now().Add(d), done: c}
	}
}

func (s *ptimeWindow) ShouldClose(in message.Message, err error) bool {
	return time.Now().After(s.end)
}

func (s *ptimeWindow) AsyncClosingChannel() <-chan struct{} {
	return s.done
}
