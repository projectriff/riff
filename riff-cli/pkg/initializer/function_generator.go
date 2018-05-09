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

func createFunction(opts options.InitOptions, functionTemplate projectriff_v1.Function) ([]byte, error) {
	function := functionTemplate.DeepCopy()
	function.TypeMeta.APIVersion = apiVersion
	function.TypeMeta.Kind = functionKind
	function.ObjectMeta.Name = opts.FunctionName
	function.Spec.Input = opts.Input
	function.Spec.Output = opts.Output
	function.Spec.Container.Image = options.ImageName(opts)
	if opts.Protocol != "" {
		function.Spec.Protocol = opts.Protocol
	}
	if function.Spec.Windowing.IsUnbounded() {
		function.Spec.Windowing.Size = 1
	}

	bytes, err := json.Marshal(function)
	if err != nil {
		return nil, err
	}
	functionMap := map[string]interface{}{}
	err = json.Unmarshal(bytes, &functionMap)
	if err != nil {
		return nil, err
	}

	// cleanup properties we don't want to marshal
	metadata := functionMap["metadata"].(map[string]interface{})
	delete(metadata, "creationTimestamp")
	spec := functionMap["spec"].(map[string]interface{})
	container := spec["container"].(map[string]interface{})
	if container["name"] == "" {
		delete(container, "name")
	}
	if len(container["resources"].(map[string]interface{})) == 0 {
		delete(container, "resources")
	}

	return yaml.Marshal(functionMap)
}

func createFunctionYaml(functionTemplate projectriff_v1.Function, opts options.InitOptions) (string, error) {
	var buffer bytes.Buffer

	function, err := createFunction(opts, functionTemplate)
	if err != nil {
		return "", err
	}

	buffer.WriteString("---\n")
	buffer.Write(function)

	return buffer.String(), nil
}
