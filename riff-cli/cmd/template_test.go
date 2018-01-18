/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTopics (t *testing.T) {
	as := assert.New(t)

	opts := InitOptions{functionName: "myfunc", input: "in", output: "out"}
	err := CreateTopics(".", opts)

	as.NoError(err)
}

func TestFunction (t *testing.T) {
	as := assert.New(t)

	opts := InitOptions{functionName: "myfunc", input: "in", output: "out", protocol:"http"}
	err := createFunction(".", "projectriff/myfunc:0.0.1", opts)

	as.NoError(err)
}