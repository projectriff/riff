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
	"path/filepath"
	"bytes"
	"errors"
	"text/template"
	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/pkg/osutils"
)

//TODO: Enable custom templates
var pythonFunctionDockerfileTemplate = `
FROM projectriff/python2-function-invoker:{{.RiffVersion}}
ARG FUNCTION_MODULE={{.ArtifactBase}}
ARG FUNCTION_HANDLER={{.Handler}}
ADD ./{{.ArtifactBase}} /
{{- if .RequirementsTextExists }}
ADD ./requirements.txt /
RUN  pip install --upgrade pip && pip install -r /requirements.txt
{{ end -}}
ENV FUNCTION_URI file:///${FUNCTION_MODULE}?handler=${FUNCTION_HANDLER}
`
var nodeFunctionDockerfileTemplate = `
FROM projectriff/node-function-invoker:{{.RiffVersion}}
ENV FUNCTION_URI /functions/{{.Artifact}}
ADD {{.ArtifactBase}} ${FUNCTION_URI}
`
var javaFunctionDockerfileTemplate = `
FROM projectriff/java-function-invoker:{{.RiffVersion}}
ARG FUNCTION_JAR=/functions/{{.ArtifactBase}}
ARG FUNCTION_CLASS={{.Handler}}
ADD target/{{.ArtifactBase}} $FUNCTION_JAR
ENV FUNCTION_URI file://${FUNCTION_JAR}?handler=${FUNCTION_CLASS}
`
var shellFunctionDockerfileTemplate = `
FROM projectriff/shell-function-invoker:{{.RiffVersion}}
ARG FUNCTION_URI="/{{.ArtifactBase}}"
ADD {{.Artifact}} /
ENV FUNCTION_URI $FUNCTION_URI
`

type DockerFileTokens struct {
	Artifact     string
	ArtifactBase string
	RiffVersion  string
	Handler      string
}

type PythonDockerFileTokens struct {
	DockerFileTokens
	RequirementsTextExists bool
}

func generateDockerfile(language string, opts options.InitOptions) (string, error) {
	switch language {
	case "java":
		return generateJavaFunctionDockerFile(opts)
	case "python":
		return generatePythonFunctionDockerFile(opts)
	case "shell":
		return generateShellFunctionDockerFile(opts)
	case "node":
		return generateNodeFunctionDockerFile(opts)
	case "js":
		return generateNodeFunctionDockerFile(opts)
	}
	return "", errors.New(fmt.Sprintf("unsupported language %s", language))
}

func generateShellFunctionDockerFile(opts options.InitOptions) (string, error) {
	dockerFileTokens := DockerFileTokens{
		Artifact:     opts.Artifact,
		ArtifactBase: filepath.Base(opts.Artifact),
		RiffVersion:  opts.RiffVersion,
	}
	return generateFunctionDockerFileContents(shellFunctionDockerfileTemplate, "docker-shell", dockerFileTokens)
}

func generateNodeFunctionDockerFile(opts options.InitOptions) (string, error) {
	dockerFileTokens := DockerFileTokens{
		Artifact:     opts.Artifact,
		ArtifactBase: filepath.Base(opts.Artifact),
		RiffVersion:  opts.RiffVersion,
	}
	return generateFunctionDockerFileContents(nodeFunctionDockerfileTemplate, "docker-node", dockerFileTokens)
}

func generateJavaFunctionDockerFile(opts options.InitOptions) (string, error) {
	dockerFileTokens := DockerFileTokens{
		Artifact:     opts.Artifact,
		ArtifactBase: filepath.Base(opts.Artifact),
		RiffVersion:  opts.RiffVersion,
		Handler:      opts.Handler,
	}
	return generateFunctionDockerFileContents(javaFunctionDockerfileTemplate, "docker-java", dockerFileTokens)
}

func generatePythonFunctionDockerFile(opts options.InitOptions) (string, error) {
	dockerFileTokens := PythonDockerFileTokens{}
	dockerFileTokens.Artifact = opts.Artifact
	dockerFileTokens.ArtifactBase = filepath.Base(opts.Artifact)
	dockerFileTokens.RiffVersion = opts.RiffVersion
	dockerFileTokens.Handler = opts.Handler
	dockerFileTokens.RequirementsTextExists = requirementTextExists(opts.FunctionPath)

	return generateFunctionDockerFileContents(pythonFunctionDockerfileTemplate, "docker-python", dockerFileTokens)
}

func requirementTextExists(functionPath string) bool {
	if !osutils.IsDirectory(functionPath) {
		functionPath = filepath.Dir(functionPath)
	}
	return osutils.FileExists(filepath.Join(functionPath, "requirements.txt"))
}

func generateFunctionDockerFileContents(tmpl string, name string, tokens interface{}) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, tokens)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
