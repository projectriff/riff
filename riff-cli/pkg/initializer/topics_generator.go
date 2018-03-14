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
	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createTopic(name string, topicSpec projectriff_v1.TopicSpec) ([]byte, error) {
	topic := projectriff_v1.Topic{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       topicKind,
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
		},
		Spec: *topicSpec.DeepCopy(),
	}

	bytes, err := json.Marshal(topic)
	if err != nil {
		return nil, err
	}
	topicMap := map[string]interface{}{}
	err = json.Unmarshal(bytes, &topicMap)
	if err != nil {
		return nil, err
	}

	// cleanup properties we don't want to marshal
	metadata := topicMap["metadata"].(map[string]interface{})
	delete(metadata, "creationTimestamp")
	if len(topicMap["spec"].(map[string]interface{})) == 0 {
		delete(topicMap, "spec")
	}

	return yaml.Marshal(topicMap)
}

func createTopicsYaml(topicSpec projectriff_v1.TopicSpec, opts options.InitOptions) (string, error) {
	var buffer bytes.Buffer

	inputTopic, err := createTopic(opts.Input, topicSpec)
	if err != nil {
		return "", err
	}

	buffer.WriteString("---\n")
	buffer.Write(inputTopic)

	if opts.Output == "" {
		return buffer.String(), nil
	}

	outputTopic, err := createTopic(opts.Output, topicSpec)
	if err != nil {
		return "", err
	}

	buffer.WriteString("\n")
	buffer.WriteString("---\n")
	buffer.Write(outputTopic)

	return buffer.String(), nil
}
