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
	"text/template"
)

type DockerFileTokens struct {
	Artifact     string
	ArtifactBase string
	RiffVersion  string
	Handler      string
}

func GenerateFunctionDockerFileContents(tmpl string, name string, tokens interface{}) (string, error) {
	return generateContents(tmpl, name, tokens)
}

func GenerateFunctionDockerIgnoreContents(tmpl string, name string, tokens interface{}) (string, error) {
	return generateContents(tmpl, name, tokens)
}

func generateContents(tmpl string, name string, tokens interface{}) (string, error) {
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
