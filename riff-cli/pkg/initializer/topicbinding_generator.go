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

func createTopicBinding(opts options.InitOptions, topicBindingTemplate projectriff_v1.TopicBinding) ([]byte, error) {
	topicBinding := topicBindingTemplate.DeepCopy()
	topicBinding.TypeMeta.APIVersion = apiVersion
	topicBinding.TypeMeta.Kind = topicBindingKind
	topicBinding.ObjectMeta.Name = opts.FunctionName
	topicBinding.Spec.Function = opts.FunctionName
	topicBinding.Spec.Input = opts.Input
	topicBinding.Spec.Output = opts.Output

	bytes, err := json.Marshal(topicBinding)
	if err != nil {
		return nil, err
	}
	topicBindingMap := map[string]interface{}{}
	err = json.Unmarshal(bytes, &topicBindingMap)
	if err != nil {
		return nil, err
	}

	// cleanup properties we don't want to marshal
	metadata := topicBindingMap["metadata"].(map[string]interface{})
	delete(metadata, "creationTimestamp")

	return yaml.Marshal(topicBindingMap)
}

func createTopicBindingYaml(topicBindingTemplate projectriff_v1.TopicBinding, opts options.InitOptions) (string, error) {
	var buffer bytes.Buffer

	topicBinding, err := createTopicBinding(opts, topicBindingTemplate)
	if err != nil {
		return "", err
	}

	buffer.WriteString("---\n")
	buffer.Write(topicBinding)

	return buffer.String(), nil
}
