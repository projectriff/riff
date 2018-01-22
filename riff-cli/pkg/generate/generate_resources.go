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
	"text/template"
	"bytes"
	"github.com/projectriff/riff-cli/pkg/options"
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



//TODO: Flag for number of partitions?
func createTopics(opts options.InitOptions) (string, error) {

	var topicTemplate = `
apiVersion : {{.ApiVersion}}
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

	input := Topic{ApiVersion: apiVersion, Name: opts.Input, Partitions: 1}
	err = tmpl.Execute(&buffer, input)
	if err != nil {
		return "", err
	}
	if opts.Output != "" {
		output := Topic{ApiVersion: apiVersion, Name: opts.Output, Partitions: 1}
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

func createFunction(opts options.InitOptions) (string, error) {
	function := Function{
		ApiVersion: apiVersion,
		Name:       opts.FunctionName,
		Input:      opts.Input,
		Output:     opts.Output,
		Protocol:   opts.Protocol,
		Image:      options.ImageName(opts)}

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

