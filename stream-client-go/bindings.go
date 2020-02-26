/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"io/ioutil"
	"path/filepath"
)

// NewStreamClientFromBinding constructs a StreamClient by reading the relevant configuration values from
// a binding directory structure.
func NewStreamClientFromBinding(path string) (*StreamClient, error) {
	var gateway, topic, contentType string
	if bytes, err := ioutil.ReadFile(filepath.Join(path, "secret", "gateway")) ; err != nil {
		return nil, err
	} else {
		gateway = string(bytes)
	}
	if bytes, err := ioutil.ReadFile(filepath.Join(path, "secret", "topic")) ; err != nil {
		return nil, err
	} else {
		topic = string(bytes)
	}
	if bytes, err := ioutil.ReadFile(filepath.Join(path, "metadata", "contentType")) ; err != nil {
		return nil, err
	} else {
		contentType = string(bytes)
	}
	return NewStreamClient(gateway, topic, contentType)
}
