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

package initializer

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"
	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/pkg/options"
)

func createLink(opts options.InitOptions, linkTemplate projectriff_v1.Link) ([]byte, error) {
	link := linkTemplate.DeepCopy()
	link.TypeMeta.APIVersion = apiVersion
	link.TypeMeta.Kind = linkKind
	link.ObjectMeta.Name = opts.FunctionName
	link.Spec.Function = opts.FunctionName
	link.Spec.Input = opts.Input
	link.Spec.Output = opts.Output

	bytes, err := json.Marshal(link)
	if err != nil {
		return nil, err
	}
	linkMap := map[string]interface{}{}
	err = json.Unmarshal(bytes, &linkMap)
	if err != nil {
		return nil, err
	}

	// cleanup properties we don't want to marshal
	metadata := linkMap["metadata"].(map[string]interface{})
	delete(metadata, "creationTimestamp")

	return yaml.Marshal(linkMap)
}

func createLinkYaml(linkTemplate projectriff_v1.Link, opts options.InitOptions) (string, error) {
	var buffer bytes.Buffer

	link, err := createLink(opts, linkTemplate)
	if err != nil {
		return "", err
	}

	buffer.WriteString("---\n")
	buffer.Write(link)

	return buffer.String(), nil
}
