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

package python

import (
	"github.com/projectriff/riff-cli/pkg/options"
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPythonDockerfile(t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:     "demo.py",
		RiffVersion:  "0.0.3",
		FunctionPath: "test_dir/python/demo",
		Handler:      "process",
	}

	docker, err := generatePythonFunctionDockerFile(opts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/python2-function-invoker:%s", opts.RiffVersion))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_MODULE=%s", opts.Artifact))
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_HANDLER=%s", opts.Handler))
	as.Contains(docker, fmt.Sprintf("ADD ./%s /", opts.Artifact))
	as.NotContains(docker, "requirements.txt")
	as.NotContains(docker, "pip")
}
