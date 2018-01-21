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
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/projectriff/riff-cli/pkg/options"
)

func TestJavaDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:     "target/greeter-1.0.0.jar",
		RiffVersion:   "0.0.2",
	}

	haOpts := options.HandlerAwareInitOptions{
		InitOptions:opts,
		Handler:"functions.Greeter",
	}

	docker, err := generateDockerfile("java",haOpts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/java-function-invoker:%s",opts.RiffVersion))
	as.Contains(docker, "ARG FUNCTION_JAR=/functions/greeter-1.0.0.jar")
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_CLASS=%s",haOpts.Handler))
	as.Contains(docker, fmt.Sprintf("ADD %s $FUNCTION_JAR",opts.Artifact))
}

func TestPythonDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:     "demo.py",
		RiffVersion:   "0.0.3",
	}

	haOpts := options.HandlerAwareInitOptions{
		InitOptions:opts,
		Handler:"process",
	}

	docker, err := generateDockerfile("python",haOpts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/python2-function-invoker:%s",opts.RiffVersion))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_MODULE=%s",opts.Artifact))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_HANDLER=%s",haOpts.Handler))
	as.Contains(docker, fmt.Sprintf("ADD ./%s /",opts.Artifact))
}

func TestNodeDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:     "square.js",
		RiffVersion:   "0.0.3",
	}

	haOpts := options.HandlerAwareInitOptions{
		InitOptions:opts,
		Handler:"process",
	}

	docker, err := generateDockerfile("node",haOpts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/node-function-invoker:%s",opts.RiffVersion))
	as.Contains(docker, fmt.Sprintf("ENV FUNCTION_URI /functions/%s",opts.Artifact))
	as.Contains(docker, fmt.Sprintf("ADD %s ${FUNCTION_URI}",opts.Artifact))
}

func TestShellDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:     "echo.sh",
		RiffVersion:   "0.0.1-snapshot",
	}

	haOpts := options.HandlerAwareInitOptions{
		InitOptions:opts,
		Handler:"process",
	}

	docker, err := generateDockerfile("shell",haOpts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/shell-function-invoker:%s",opts.RiffVersion))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_URI=\"/%s\"",opts.Artifact))
	as.Contains(docker, fmt.Sprintf("ADD %s /",opts.Artifact))
}
