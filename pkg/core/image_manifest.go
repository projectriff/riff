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

	"github.com/ghodss/yaml"
)

const imageManifestVersion_0_1 = "0.1"

// ImageManifest defines the image names found in YAML files of system components.
type ImageManifest struct {
	ManifestVersion string `json:"manifestVersion"`
	Images          []string
}

func NewImageManifest(path string) (*ImageManifest, error) {
	var m ImageManifest
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading image manifest file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &m)
	if err != nil {
		return nil, fmt.Errorf("error parsing image manifest file: %v", err)
	}

	if m.ManifestVersion != imageManifestVersion_0_1 {
		return nil, fmt.Errorf("image manifest has unsupported version: %s", m.ManifestVersion)
	}

	if m.Images == nil {
		return nil, fmt.Errorf("image manifest is incomplete: images array is missing: %#v", m)
	}

	return &m, nil
}
