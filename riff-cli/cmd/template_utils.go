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
	"os"
	"bytes"
)

const (
	apiVersion = "projectriff.io/v1"
)

type Topic struct {
	ApiVersion string
	Name string
	Partitions int
}

type Function struct {
	ApiVersion string
	Name string
	Input string
	Output string
	Image string
	Protocol string
}

//TODO: Flag for number of partitions?
func CreateTopics(workDir string, opts InitOptions) error {


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
			return err
		}

		input := Topic{ApiVersion:apiVersion, Name: opts.input, Partitions: 1}

		output := Topic{ApiVersion:apiVersion, Name: opts.output, Partitions: 1}

		err = tmpl.Execute(os.Stdout, input)
		if err != nil {
			return err
		}

		err = tmpl.Execute(os.Stdout, output)
		return err

}

func createFunction(workDir string, image string, opts InitOptions) error {
	var functionTemplate = `
apiVersion: {{.ApiVersion}}
kind: Function
metadata:
  name: {{.Name}}
spec:
  protocol: {{.Protocol}}
  input: {{.Input}}
  container:
    image: {{.Image}}
`
	var functionWithOutputTemplate = `
apiVersion: {{.ApiVersion}}
kind: Function
metadata:
  name: {{.Name}}
spec:
  protocol: {{.Protocol}}
  input: {{.Input}}
  output: {{.Output}}
  container:
    image: {{.Image}}
`

	function := Function{ApiVersion:apiVersion, Name:opts.functionName, Input:opts.input, Output:opts.output, Protocol:opts.protocol, Image:image}

	var tmpl *template.Template
	var err error
	if function.Output == "" {
		tmpl, err = template.New("function").Parse(functionTemplate)
		if err != nil {
			return err
		}
	} else {
		tmpl, err = template.New("function").Parse(functionWithOutputTemplate)
		if err != nil {
			return err
		}
	}
	err = tmpl.Execute(os.Stdout, function)
	return err
}

type LongVals struct {
	Process string
	Command string
	Result  string
}

func createCmdLong(longDescr string,  vals LongVals) string {
	tmpl, err := template.New("longDescr").Parse(longDescr)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, vals)
	if err != nil {
		panic(err)
	}

	return tpl.String()

}
