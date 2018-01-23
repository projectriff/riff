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
		Handler:"functions.Greeter",
	}


	docker, err := generateDockerfile("java",opts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/java-function-invoker:%s",opts.RiffVersion))
	as.Contains(docker, "ARG FUNCTION_JAR=/functions/greeter-1.0.0.jar")
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_CLASS=%s",opts.Handler))
	as.Contains(docker, fmt.Sprintf("ADD %s $FUNCTION_JAR",opts.Artifact))
}

func TestPythonDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:     "demo.py",
		RiffVersion:   "0.0.3",
		FunctionPath: "test_dir/python/demo",
		Handler:"process",
	}

	docker, err := generateDockerfile("python",opts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/python2-function-invoker:%s",opts.RiffVersion))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_MODULE=%s",opts.Artifact))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_HANDLER=%s",opts.Handler))
	as.Contains(docker, fmt.Sprintf("ADD ./%s /",opts.Artifact))
	as.NotContains(docker, "requirements.txt")
	as.NotContains(docker, "pip")
}

func TestPythonDockerfileWithDeps (t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:     "demo.py",
		RiffVersion:   "0.0.3",
		FunctionPath: "../../test_data/python/demo_with_deps",
		Handler:"process",
	}

	docker, err := generateDockerfile("python",opts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/python2-function-invoker:%s",opts.RiffVersion))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_MODULE=%s",opts.Artifact))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_HANDLER=%s",opts.Handler))
	as.Contains(docker, fmt.Sprintf("ADD ./%s /",opts.Artifact))
	as.Contains(docker, "requirements.txt")
	as.Contains(docker, "pip")
}

func TestNodeDockerfile (t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:     "square.js",
		RiffVersion:   "0.0.3",
		Handler:"process",
	}

	docker, err := generateDockerfile("node",opts)
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
		Handler:"process",
	}

	docker, err := generateDockerfile("shell",opts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/shell-function-invoker:%s",opts.RiffVersion))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_URI=\"/%s\"",opts.Artifact))
	as.Contains(docker, fmt.Sprintf("ADD %s /",opts.Artifact))
}
