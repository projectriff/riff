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

func createBinding(opts options.InitOptions, bindingTemplate projectriff_v1.Binding) ([]byte, error) {
	binding := bindingTemplate.DeepCopy()
	binding.TypeMeta.APIVersion = apiVersion
	binding.TypeMeta.Kind = bindingKind
	binding.ObjectMeta.Name = opts.FunctionName
	binding.Spec.Function = opts.FunctionName
	binding.Spec.Input = opts.Input
	binding.Spec.Output = opts.Output

	bytes, err := json.Marshal(binding)
	if err != nil {
		return nil, err
	}
	bindingMap := map[string]interface{}{}
	err = json.Unmarshal(bytes, &bindingMap)
	if err != nil {
		return nil, err
	}

	// cleanup properties we don't want to marshal
	metadata := bindingMap["metadata"].(map[string]interface{})
	delete(metadata, "creationTimestamp")

	return yaml.Marshal(bindingMap)
}

func createBindingYaml(bindingTemplate projectriff_v1.Binding, opts options.InitOptions) (string, error) {
	var buffer bytes.Buffer

	binding, err := createBinding(opts, bindingTemplate)
	if err != nil {
		return "", err
	}

	buffer.WriteString("---\n")
	buffer.Write(binding)

	return buffer.String(), nil
}
