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

package generate

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/projectriff/riff-cli/pkg/options"
	"gopkg.in/yaml.v2"
)

func TestTopics(t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		FunctionName: "myfunc",
		Input:        "in",
		Output:       "out",
	}
	topic, err := createTopics(opts)

	as.NoError(err)
	as.Contains(topic, "name: in")
	as.Contains(topic, "name: out")
}


type YFunction struct {
	ApiVersion string
	Kind string
	Metadata struct {
		Name string
	}
	Spec struct {
		Protocol string
		Input string
		Container struct {
			Image string
		}
	}
}

func TestFunction(t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		FunctionName: "myfunc",
		Input:        "in",
		Output:       "out",
		Protocol:     "http",
		UserAccount:   "me",
		Version:       "0.0.1",
	}

	f, err := createFunction(opts)
	as.NoError(err)
	as.Contains(f, "input:")
	as.Contains(f, "output:")

	opts.Output = ""

	f, err = createFunction(opts)
	as.NoError(err)

	yf := YFunction{}
	err = yaml.Unmarshal([]byte(f), &yf)
	as.NoError(err)
	as.Equal("http",yf.Spec.Protocol)
	as.Equal("in",yf.Spec.Input)
	as.Equal("myfunc",yf.Metadata.Name)
	as.Equal("me/myfunc:0.0.1",yf.Spec.Container.Image)
}
