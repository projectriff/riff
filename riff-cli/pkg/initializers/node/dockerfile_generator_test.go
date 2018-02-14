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

package node

import (
	"fmt"
	"testing"

	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/stretchr/testify/assert"
)

func TestNodeDockerfile(t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:    "square.js",
		RiffVersion: "0.0.3",
		FilePath:    "../../../test_data/node/square/square.js",
		Handler:     "process",
	}

	docker, err := generateNodeFunctionDockerFile(opts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/node-function-invoker:%s\n", opts.RiffVersion))
	as.Contains(docker, fmt.Sprintf("ENV FUNCTION_URI /functions/%s\n", opts.Artifact))
	as.Contains(docker, fmt.Sprintf("ADD %s ${FUNCTION_URI}\n", opts.Artifact))

	as.NotContains(docker, "ENV FUNCTION_URI /functions/\n")
	as.NotContains(docker, "COPY . ${FUNCTION_URI}\n")
	as.NotContains(docker, "RUN (cd ${FUNCTION_URI} && npm install --production)\n")
}

func TestNodePackageDockerfile(t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:    "square.js",
		RiffVersion: "0.0.3",
		FilePath:    "../../../test_data/node/square-package/square.js",
		Handler:     "process",
	}

	docker, err := generateNodeFunctionDockerFile(opts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/node-function-invoker:%s\n", opts.RiffVersion))
	as.Contains(docker, "ENV FUNCTION_URI /functions/\n")
	as.Contains(docker, "COPY . ${FUNCTION_URI}\n")
	as.Contains(docker, "RUN (cd ${FUNCTION_URI} && npm install --production)\n")

	as.NotContains(docker, fmt.Sprintf("ENV FUNCTION_URI /functions/%s\n", opts.Artifact))
	as.NotContains(docker, fmt.Sprintf("ADD %s ${FUNCTION_URI}\n", opts.Artifact))
}
