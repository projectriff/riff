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

package server

import (
	"sync"

	"github.com/projectriff/message-transport/pkg/message"
)

// Type repliesMap implements a concurrent safe map of channels to send replies to, keyed by message correlationIds
type repliesMap struct {
	m    map[string]chan<- message.Message
	lock sync.RWMutex
}

func (replies *repliesMap) Delete(key string) {
	replies.lock.Lock()
	defer replies.lock.Unlock()
	delete(replies.m, key)
}

func (replies *repliesMap) Get(key string) chan<- message.Message {
	replies.lock.RLock()
	defer replies.lock.RUnlock()
	return replies.m[key]
}

func (replies *repliesMap) Put(key string, value chan<- message.Message) {
	replies.lock.Lock()
	defer replies.lock.Unlock()
	replies.m[key] = value
}

func newRepliesMap() *repliesMap {
	return &repliesMap{make(map[string]chan<- message.Message), sync.RWMutex{}}
}
