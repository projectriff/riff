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

package core

import (
	"bytes"
	"github.com/projectriff/riff-cli/pkg/options"
	"text/template"
)

type Topic struct {
	ApiVersion string
	Name       string
	Partitions int
}

//TODO: Flag for number of partitions?
func createTopics(opts options.InitOptions) (string, error) {

	var topicTemplate string = `
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

	input := Topic{ApiVersion: ApiVersion, Name: opts.Input, Partitions: 1}
	err = tmpl.Execute(&buffer, input)
	if err != nil {
		return "", err
	}
	if opts.Output != "" {
		output := Topic{ApiVersion: ApiVersion, Name: opts.Output, Partitions: 1}
		err = tmpl.Execute(&buffer, output)
		if err != nil {
			return "", err
		}
	}

	return buffer.String(), nil
}
