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
	"fmt"
)

func TestTopics (t *testing.T) {
	as := assert.New(t)

	opts := InitOptions{functionName: "myfunc", input: "in", output: "out"}
	topic, err := createTopics(opts)

	as.NoError(err)
	as.Contains(topic, "name: in")
	as.Contains(topic, "name: out")
}

func TestFunction (t *testing.T) {
	as := assert.New(t)

	opts := InitOptions{functionName: "myfunc", input: "in", output: "out", protocol:"http"}
	f, err := createFunction(opts)
	as.NoError(err)
	as.Contains(f, "input:")
	as.Contains(f, "output:")

	opts = InitOptions{functionName: "myfunc", input: "in", protocol:"http"}
	f, err = createFunction(opts)
	as.NoError(err)
	as.Contains(f, "input:")
	as.NotContains(f, "output:")
}

func TestJavaDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := InitOptions{
		artifact:     "target/greeter-1.0.0.jar",
		riffVersion:   RIFF_VERSION,
	}

	haOpts := HandlerAwareInitOptions{
		InitOptions:opts,
		handler:"functions.Greeter",
	}

	docker, err := createDockerfile("java",haOpts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/java-function-invoker:%s",opts.riffVersion))
	as.Contains(docker, "ARG FUNCTION_JAR=/functions/greeter-1.0.0.jar")
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_CLASS=%s",haOpts.handler))
	as.Contains(docker, fmt.Sprintf("ADD %s $FUNCTION_JAR",opts.artifact))
}

func TestPythonDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := InitOptions{
		artifact:     "demo.py",
		riffVersion:   "0.0.3",
	}

	haOpts := HandlerAwareInitOptions{
		InitOptions:opts,
		handler:"process",
	}

	docker, err := createDockerfile("python",haOpts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/python2-function-invoker:%s",opts.riffVersion))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_MODULE=%s",opts.artifact))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_HANDLER=%s",haOpts.handler))
	as.Contains(docker, fmt.Sprintf("ADD ./%s /",opts.artifact))
}

func TestNodeDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := InitOptions{
		artifact:     "square.js",
		riffVersion:   "0.0.3",
	}

	haOpts := HandlerAwareInitOptions{
		InitOptions:opts,
		handler:"process",
	}

	docker, err := createDockerfile("node",haOpts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/node-function-invoker:%s",opts.riffVersion))
	as.Contains(docker, fmt.Sprintf("ENV FUNCTION_URI /functions/%s",opts.artifact))
	as.Contains(docker, fmt.Sprintf("ADD %s ${FUNCTION_URI}",opts.artifact))
}

func TestShellDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := InitOptions{
		artifact:     "echo.sh",
		riffVersion:   RIFF_VERSION,
	}

	haOpts := HandlerAwareInitOptions{
		InitOptions:opts,
		handler:"process",
	}

	docker, err := createDockerfile("shell",haOpts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/shell-function-invoker:%s",opts.riffVersion))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_URI=\"/%s\"",opts.artifact))
	as.Contains(docker, fmt.Sprintf("ADD %s /",opts.artifact))
}