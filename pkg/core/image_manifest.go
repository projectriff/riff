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

	"github.com/opencontainers/go-digest"

	"github.com/docker/distribution/reference"

	"github.com/ghodss/yaml"
)

const imageManifestVersion_0_1 = "0.1"

// imageName contains a full image reference in the form [host]/repository/name/parts:tag
type imageName reference.Reference

// imageDigest contains a digest of the actual image contents
type imageDigest digest.Digest

// ImageManifest defines the image names found in YAML files of system components.
type ImageManifest struct {
	ManifestVersion string                    `json:"manifestVersion"`
	Images          map[imageName]imageDigest `json:images`
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
		return nil, fmt.Errorf("image manifest is incomplete: images map is missing: %#v", m)
	}

	return &m, nil
}

func EmptyImageManifest() *ImageManifest { // Will rename to NewImageManifest once the other is renamed to LoadImageManifest
	result := &ImageManifest{}
	result.Images = make(map[imageName]imageDigest)
	result.ManifestVersion = imageManifestVersion_0_1
	return result
}

func (m *ImageManifest) Save(path string) error {
	bytes, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, outputFilePermissions)
}

func (m *ImageManifest) AddImage(i string, d string) error {
	in, err := parseImageName(i)
	if err != nil {
		return err
	}
	m.Images[in] = imageDigest(d)
	return nil
}

func (m *ImageManifest) RemoveImage(i string) error {
	in, err := parseImageName(i)
	if err != nil {
		return err
	}
	delete(m.Images, in)
	return nil
}

func parseImageName(i string) (imageName, error) {
	ref, err := reference.Parse(i)
	return imageName(ref), err
}
