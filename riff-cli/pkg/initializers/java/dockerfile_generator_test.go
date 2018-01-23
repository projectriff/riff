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

package java

import (
	"github.com/projectriff/riff-cli/pkg/options"
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestJavaDockerfile(t *testing.T) {
	as := assert.New(t)

	opts := options.InitOptions{
		Artifact:    "target/greeter-1.0.0.jar",
		RiffVersion: "0.0.2",
		Handler:     "functions.Greeter",
	}

	docker, err := generateJavaFunctionDockerFile(opts)
	as.NoError(err)
	as.Contains(docker, fmt.Sprintf("FROM projectriff/java-function-invoker:%s", opts.RiffVersion))
	as.Contains(docker, "ARG FUNCTION_JAR=/functions/greeter-1.0.0.jar")
	as.Contains(docker, fmt.Sprintf("ARG FUNCTION_CLASS=%s", opts.Handler))
	as.Contains(docker, fmt.Sprintf("ADD %s $FUNCTION_JAR", opts.Artifact))
}
