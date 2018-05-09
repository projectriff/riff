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

package initializer

import (
	"testing"

	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/stretchr/testify/assert"
)

func TestFunctionYaml(t *testing.T) {
	as := assert.New(t)

	functionTemplate := projectriff_v1.Function{
		Spec: projectriff_v1.FunctionSpec{
			Protocol: "grpc",
		},
	}
	opts := options.InitOptions{
		FunctionName: "myfunc",
		Input:        "in",
		UserAccount:  "me",
		Version:      "0.0.1",
	}
	yaml, err := createFunctionYaml(functionTemplate, opts)

	t.Log(yaml)

	as.NoError(err)
	as.Equal(yaml, `---
apiVersion: projectriff.io/v1alpha1
kind: Function
metadata:
  name: myfunc
spec:
  container:
    image: me/myfunc:0.0.1
  input: in
  protocol: grpc
  windowing:
    size: 1
`)
}

func TestFunctionYaml_WithOutput(t *testing.T) {
	as := assert.New(t)

	functionTemplate := projectriff_v1.Function{
		Spec: projectriff_v1.FunctionSpec{
			Protocol: "grpc",
		},
	}
	opts := options.InitOptions{
		FunctionName: "myfunc",
		Input:        "in",
		Output:       "out",
		UserAccount:  "me",
		Version:      "0.0.1",
		Protocol:     "http",
	}
	yaml, err := createFunctionYaml(functionTemplate, opts)

	t.Log(yaml)

	as.NoError(err)
	as.Equal(yaml, `---
apiVersion: projectriff.io/v1alpha1
kind: Function
metadata:
  name: myfunc
spec:
  container:
    image: me/myfunc:0.0.1
  input: in
  output: out
  protocol: http
  windowing:
    size: 1
`)
}

func TestFunctionYaml_WithProtocolOverride(t *testing.T) {
	as := assert.New(t)

	functionTemplate := projectriff_v1.Function{
		Spec: projectriff_v1.FunctionSpec{
			Protocol: "grpc",
		},
	}
	opts := options.InitOptions{
		FunctionName: "myfunc",
		Input:        "in",
		UserAccount:  "me",
		Version:      "0.0.1",
		Protocol:     "http",
	}
	yaml, err := createFunctionYaml(functionTemplate, opts)

	t.Log(yaml)

	as.NoError(err)
	as.Equal(yaml, `---
apiVersion: projectriff.io/v1alpha1
kind: Function
metadata:
  name: myfunc
spec:
  container:
    image: me/myfunc:0.0.1
  input: in
  protocol: http
  windowing:
    size: 1
`)
}
