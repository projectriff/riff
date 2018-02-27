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
	"github.com/projectriff/riff-cli/pkg/initializers/core"
	"github.com/projectriff/riff-cli/pkg/options"
	"path/filepath"
	"github.com/projectriff/riff-cli/pkg/osutils"
)

type PythonDockerFileTokens struct {
	core.DockerFileTokens
	RequirementsTextExists bool
}

var pythonFunctionDockerfileTemplate = `
FROM projectriff/python2-function-invoker:{{.RiffVersion}}
ARG FUNCTION_MODULE={{.ArtifactBase}}
ARG FUNCTION_HANDLER={{.Handler}}
ADD ./{{.ArtifactBase}} /
{{- if .RequirementsTextExists }}
ADD ./requirements.txt /
RUN  pip install --upgrade pip && pip install -r /requirements.txt
{{- end }}
ENV FUNCTION_URI file:///${FUNCTION_MODULE}?handler=${FUNCTION_HANDLER}
`

func generatePythonFunctionDockerFile(opts options.InitOptions) (string, error) {
	dockerFileTokens := PythonDockerFileTokens{}
	dockerFileTokens.Artifact = opts.Artifact
	dockerFileTokens.ArtifactBase = filepath.Base(opts.Artifact)
	dockerFileTokens.RiffVersion = opts.RiffVersion
	dockerFileTokens.Handler = opts.Handler
	dockerFileTokens.RequirementsTextExists = requirementTextExists(opts.FilePath)

	return core.GenerateFunctionDockerFileContents(pythonFunctionDockerfileTemplate, "docker-python", dockerFileTokens)
}

func requirementTextExists(filePath string) bool {
	if !osutils.IsDirectory(filePath) {
		filePath = filepath.Dir(filePath)
	}
	return osutils.FileExists(filepath.Join(filePath, "requirements.txt"))
}