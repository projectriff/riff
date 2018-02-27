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
	"path/filepath"

	"github.com/projectriff/riff-cli/pkg/initializers/core"
	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/pkg/osutils"
)

type NodeDockerFileTokens struct {
	core.DockerFileTokens
	PackageJSONExists bool
}

var nodeFunctionDockerfileTemplate = `
FROM projectriff/node-function-invoker:{{.RiffVersion}}
{{ if .PackageJSONExists -}}
ENV FUNCTION_URI /functions/
COPY . ${FUNCTION_URI}
RUN (cd ${FUNCTION_URI} && npm install --production)
{{- else -}}
ENV FUNCTION_URI /functions/{{.Artifact}}
ADD {{.ArtifactBase}} ${FUNCTION_URI}
{{- end }}
`

var nodeFunctionDockerIgnoreTemplate = `
{{ if .PackageJSONExists -}}
node_modules
{{- end }}
`

func generateNodeFunctionDockerFile(opts options.InitOptions) (string, error) {
	dockerFileTokens := generateDockerFileTokens(opts)
	return core.GenerateFunctionDockerFileContents(nodeFunctionDockerfileTemplate, "docker-node", dockerFileTokens)
}

func generateNodeFunctionDockerIgnore(opts options.InitOptions) (string, error) {
	dockerFileTokens := generateDockerFileTokens(opts)
	return core.GenerateFunctionDockerIgnoreContents(nodeFunctionDockerIgnoreTemplate, "docker-node", dockerFileTokens)
}

func generateDockerFileTokens(opts options.InitOptions) NodeDockerFileTokens {
	dockerFileTokens := NodeDockerFileTokens{}
	dockerFileTokens.Artifact = opts.Artifact
	dockerFileTokens.ArtifactBase = filepath.Base(opts.Artifact)
	dockerFileTokens.RiffVersion = opts.RiffVersion
	dockerFileTokens.PackageJSONExists = packageJSONExists(opts.FilePath)
	return dockerFileTokens
}

func packageJSONExists(filePath string) bool {
	if !osutils.IsDirectory(filePath) {
		filePath = filepath.Dir(filePath)
	}
	return osutils.FileExists(filepath.Join(filePath, "package.json"))
}
