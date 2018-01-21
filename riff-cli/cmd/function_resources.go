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
	"text/template"
	"bytes"
	"path/filepath"
	"fmt"
	"errors"
)

const (
	apiVersion = "projectriff.io/v1"
)

type FunctionResources struct {
	Topics     string
	Function   string
	DockerFile string
}

type Topic struct {
	ApiVersion string
	Name       string
	Partitions int
}

type Function struct {
	ApiVersion string
	Name       string
	Input      string
	Output     string
	Image      string
	Protocol   string
}

func createFunctionResources(workDir, language string, opts HandlerAwareInitOptions) error {


	var functionResources FunctionResources
	var err error
	functionResources.Topics, err = createTopics(opts.InitOptions)
	if err != nil {
		return err
	}
	functionResources.Function, err = createFunction(opts.InitOptions)
	if err != nil {
		return err
	}
	functionResources.DockerFile, err = createDockerfile(language,opts)
	if err != nil {
		return err
	}

	if (opts.dryRun) {
		fmt.Println("Generated Topics:\n")
		fmt.Printf("%s\n",functionResources.Topics)
		fmt.Println("\nGenerated Function:\n")
		fmt.Printf("%s\n",functionResources.Function)
		fmt.Println("\nGenerated Dockerfile:\n")
		fmt.Printf("%s\n",functionResources.DockerFile)
	} else {
		//Write Files
	}
	return nil
}

//TODO: Flag for number of partitions?
func createTopics(opts InitOptions) (string, error) {

	var topicTemplate = `
apiversion : {{.ApiVersion}}
kind: Topic
metadata:	
	name: {{.Name}}
spec:
	partitions: {{.Partitions}}
`
	tmpl, err := template.New("topic").Parse(topicTemplate)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer

	input := Topic{ApiVersion: apiVersion, Name: opts.input, Partitions: 1}
	err = tmpl.Execute(&buffer, input)
	if err != nil {
		return "", err
	}
	if opts.output != "" {
		output := Topic{ApiVersion: apiVersion, Name: opts.output, Partitions: 1}
		err = tmpl.Execute(&buffer, output)
		if err != nil {
			return "", err
		}
	}

	return buffer.String(), nil
}

//TODO: Kludgy '-' used to supress blank line, {{else}} adds a new line.
var functionTemplate = `
apiVersion: {{.ApiVersion}}
kind: Function
metadata:
  name: {{.Name}}
spec:
  protocol: {{.Protocol}}
  input: {{.Input}}
{{- if .Output}} 
  output: {{.Output}}
{{ else }}
{{ end -}}
  container:
    image: {{.Image}}
`

func createFunction(opts InitOptions) (string, error) {
	image := fmt.Sprintf("%s/%s:%s",opts.userAccount,opts.functionName,opts.version)
	function := Function{
		ApiVersion: apiVersion,
		Name:       opts.functionName,
		Input:      opts.input,
		Output:     opts.output,
		Protocol:   opts.protocol,
		Image:      image}

	var tmpl *template.Template
	var err error
	var buffer bytes.Buffer

	tmpl, err = template.New("function").Parse(functionTemplate)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&buffer, function)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

//TODO: Enable custom templates
var pythonFunctionDockerfileTemplate = `
FROM projectriff/python2-function-invoker:{{.RiffVersion}}
ARG FUNCTION_MODULE={{.ArtifactBase}}
ARG FUNCTION_HANDLER={{.Handler}}
ADD ./{{.ArtifactBase}} /
ADD ./requirements.txt /
RUN  pip install --upgrade pip && pip install -r /requirements.txt
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

func createDockerfile(language string, opts HandlerAwareInitOptions) (string, error) {
	switch language {
	case "java":
		return createJavaFunctionDockerFile(opts)
	case "python":
		return createPythonFunctionDockerFile(opts)
	case "shell":
		return createShellFunctionDockerFile(opts.InitOptions)
	case "node":
		return createNodeFunctionDockerFile(opts.InitOptions)
	case "js":
		return createNodeFunctionDockerFile(opts.InitOptions)
	}
	return "", errors.New(fmt.Sprintf("unsupported language %s", language))
}

func createShellFunctionDockerFile(opts InitOptions) (string, error) {
	dockerFileTokens := DockerFileTokens{
		Artifact:     opts.artifact,
		ArtifactBase: filepath.Base(opts.artifact),
		RiffVersion:  opts.riffVersion,
	}
	return createFunctionDockerFile(shellFunctionDockerfileTemplate, "docker-shell", dockerFileTokens)
}

func createNodeFunctionDockerFile(opts InitOptions) (string, error) {
	dockerFileTokens := DockerFileTokens{
		Artifact:     opts.artifact,
		ArtifactBase: filepath.Base(opts.artifact),
		RiffVersion:  opts.riffVersion,
	}
	return createFunctionDockerFile(nodeFunctionDockerfileTemplate, "docker-node", dockerFileTokens)
}

func createJavaFunctionDockerFile(opts HandlerAwareInitOptions) (string, error) {
	dockerFileTokens := DockerFileTokens{
		Artifact:     opts.artifact,
		ArtifactBase: filepath.Base(opts.artifact),
		RiffVersion:  opts.riffVersion,
		Handler:      opts.handler,
	}
	return createFunctionDockerFile(javaFunctionDockerfileTemplate, "docker-java", dockerFileTokens)
}

func createPythonFunctionDockerFile(opts HandlerAwareInitOptions) (string, error) {
	dockerFileTokens := DockerFileTokens{
		Artifact:     opts.artifact,
		ArtifactBase: filepath.Base(opts.artifact),
		RiffVersion:  opts.riffVersion,
		Handler:      opts.handler,
	}

	return createFunctionDockerFile(pythonFunctionDockerfileTemplate, "docker-python", dockerFileTokens)
}

func createFunctionDockerFile(tmpl string, name string, tokens DockerFileTokens) (string, error) {
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
