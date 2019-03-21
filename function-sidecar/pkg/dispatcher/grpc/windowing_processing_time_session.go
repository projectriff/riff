/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
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

// Type ptimeSession is a WindowingStrategy that delimits windows based on inactivity time.
type ptimeSession struct {
	timer    *time.Timer
	duration time.Duration
	done     chan struct{}
}

func ptimeSessionFactoryFactory(d time.Duration) WindowingStrategyFactory {
	return func(first message.Message) WindowingStrategy {
		c := make(chan struct{})
		t := time.NewTimer(d)
		go func() {
			<-time.After(d)
			close(c)
		}()
		return &ptimeSession{timer: t, done: c, duration: d}
	}
}

func (s *ptimeSession) ShouldClose(in message.Message, err error) bool {
	if !s.timer.Stop() {
		<-s.timer.C
	}
	s.timer.Reset(s.duration)
	return false
}

func (s *ptimeSession) AsyncClosingChannel() <-chan struct{} {
	return s.done
}
