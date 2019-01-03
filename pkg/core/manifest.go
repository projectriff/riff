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
 */

package core

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/projectriff/riff/pkg/fileutils"

	"github.com/ghodss/yaml"
)

const manifestVersion_0_1 = "0.1"

// Manifest defines the location of YAML files for system components.
type Manifest struct {
	ManifestVersion string   `json:"manifestVersion"`
	Istio           []string `json:"istio"`
	Knative         []string `json:"knative"`
	Namespace       []string `json:"namespace"`
	manifestDir     string
}

func ResolveManifest(manifests map[string]*Manifest, path string) (*Manifest, error) {
	if manifest, ok := manifests[path]; ok {
		return manifest, nil
	}
	return NewManifest(path)
}

func NewManifest(path string) (*Manifest, error) {
	var m Manifest
	yamlFile, err := fileutils.Read(path, "")
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

	err = checkCompleteness(m)
	if err != nil {
		return nil, err
	}

	err = m.VisitResources(checkResource)
	if err != nil {
		return nil, err
	}

	m.manifestDir, err = fileutils.Dir(path)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Manifest) VisitResources(f func(resource string) error) error {
	for _, resourceArray := range [][]string{m.Istio, m.Knative, m.Namespace} {
		for _, resource := range resourceArray {
			err := f(resource)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ResourceAbsolutePath takes a path to a resource and returns an equivalent absolute path.
// If the input path is a http(s) URL or is an absolute file path, it is returned without modification.
// If the input path is a file URL, the corresponding (absolute) file path is returned.
// If the input path is a relative file path, it is interpreted to be relative to the directory from which the
// manifest was read (and if the manifest was not read from a directory, an error is returned) and the corresponding
// absolute file path is returned.
func (m *Manifest) ResourceAbsolutePath(path string) (string, error) {
	absolute, canonicalPath, err := fileutils.IsAbsFile(path)
	if err != nil {
		return "", err
	}

	if absolute {
		return canonicalPath, nil
	}

	if m.manifestDir == "" {
		return "", errors.New("relative path undefined since manifest was not read from a directory")
	}

	return fileutils.AbsFile(path, m.manifestDir)
}

func checkCompleteness(m Manifest) error {
	var omission string
	if m.Istio == nil {
		omission = "istio"
	} else if m.Knative == nil {
		omission = "knative"
	} else if m.Namespace == nil {
		omission = "namespace"
	} else {
		return nil
	}
	return fmt.Errorf("manifest is incomplete: %s array missing: %#v", omission, m)
}

func checkResource(resource string) error {
	if filepath.IsAbs(resource) {
		return fmt.Errorf("resources must use a http or https URL or a relative path: absolute path not supported: %s", resource)
	}

	u, err := url.Parse(resource)
	if err != nil {
		return err
	}
	if u.Scheme == "http" || u.Scheme == "https" || (u.Scheme == "" && !filepath.IsAbs(u.Path)) {
		return nil
	}

	if u.Scheme == "" {
		return fmt.Errorf("resources must use a http or https URL or a relative path: absolute path not supported: %s", resource)
	}

	return fmt.Errorf("resources must use a http or https URL or a relative path: scheme %s not supported: %s", u.Scheme, resource)
}
