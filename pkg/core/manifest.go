/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package core

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const manifestVersion_0_1 = "0.1"

// Manifest defines the location of YAML files for system components.
type Manifest struct {
	ManifestVersion string `yaml:"manifestVersion"`
	Istio           []string
	Knative         []string
}

func NewManifest(path string) (*Manifest, error) {
	var m Manifest
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading manifest file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &m)
	if err != nil {
		return nil, fmt.Errorf("Error parsing manifest file: %v", err)
	}

	if m.ManifestVersion != manifestVersion_0_1 {
		return nil, fmt.Errorf("Manifest has unsupported version: %s", m.ManifestVersion)
	}

	if m.Istio == nil ||
		m.Knative == nil {
		return nil, fmt.Errorf("Manifest is incomplete: %#v", m)
	}

	return &m, nil
}
