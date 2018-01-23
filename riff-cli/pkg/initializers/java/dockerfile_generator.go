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
	"path/filepath"
	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/pkg/initializers/core"
)

var dockerfileTemplate = `
FROM projectriff/java-function-invoker:{{.RiffVersion}}
ARG FUNCTION_JAR=/functions/{{.ArtifactBase}}
ARG FUNCTION_CLASS={{.Handler}}
ADD target/{{.ArtifactBase}} $FUNCTION_JAR
ENV FUNCTION_URI file://${FUNCTION_JAR}?handler=${FUNCTION_CLASS}
`

func generateJavaFunctionDockerFile(opts options.InitOptions) (string, error) {
	dockerFileTokens := core.DockerFileTokens{
		Artifact:     opts.Artifact,
		ArtifactBase: filepath.Base(opts.Artifact),
		RiffVersion:  opts.RiffVersion,
		Handler:      opts.Handler,
	}
	return core.GenerateFunctionDockerFileContents(dockerfileTemplate, "docker-java", dockerFileTokens)
}
