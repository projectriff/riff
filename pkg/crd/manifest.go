/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package crd

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/jinzhu/copier"
	"github.com/projectriff/riff/pkg/fileutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/url"
	"path/filepath"
	"strings"
)

type ResourceChecks struct {
	Kind     string               `json:"kind,omitempty"`
	Selector metav1.LabelSelector `json:"selector,omitempty"`
	JsonPath string               `json:"jsonpath,omitempty"`
	Pattern  string               `json:"pattern,omitempty"`
}

type Resource struct {
	Path      string           `json:"path,omitempty"`
	Content   string           `json:"content,omitempty"`
	Name      string           `json:"name,omitempty"`
	Namespace string           `json:"namespace,omitempty"`
	Checks    []ResourceChecks `json:"checks,omitempty"`
}

type RiffResource struct {
	System         []Resource `json:"system,omitempty"`
	Initialization []Resource `json:"initialization,omitempty"`
}

type RiffSpec struct {
	Images    []string     `json:"images,omitempty"`
	Resources RiffResource `json:"resources,omitempty"`
}

type RiffStatus struct {
	Status string `json:"status,omitempty"`
}

type Manifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec        RiffSpec   `json:"spec,omitempty"`
	Status      RiffStatus `json:"status,omitempty"`
	manifestDir string
}

func (orig *Manifest) DeepCopyObject() runtime.Object {
	result := &Manifest{}
	copier.Copy(result, orig)
	return result
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
		if strings.Contains(err.Error(), "did not find expected key") {
			return nil, fmt.Errorf("Error parsing manifest file: %v. Please ensure that manifest has supported version:", err)
		}
		return nil, fmt.Errorf("Error parsing manifest file: %v", err)
	}

	supportedVersion := fmt.Sprintf("%s/%s", Group, Version)
	if !strings.EqualFold(m.APIVersion, supportedVersion) {
		return nil, errors.New(fmt.Sprintf("Unsupported version %s. Supported version is %s", m.APIVersion, supportedVersion))
	}

	err = checkCompleteness(m)
	if err != nil {
		return nil, err
	}

	err = m.VisitResources(checkResourcePath)
	if err != nil {
		return nil, err
	}

	m.manifestDir, err = fileutils.Dir(path)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Manifest) VisitResources(f func(res Resource) error) error {
	var resources []Resource
	resources = append(resources, m.Spec.Resources.System...)
	resources = append(resources, m.Spec.Resources.Initialization...)
	for _, resource := range resources {
		err := f(resource)
		if err != nil {
			return err
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
	if m.Spec.Resources.System == nil {
		omission = "system"
	} else if m.Spec.Resources.Initialization == nil {
		omission = "namespace-initialization"
	} else {
		return nil
	}
	return fmt.Errorf("manifest is incomplete: %s missing: %#v", omission, m)
}

func checkResourcePath(resource Resource) error {
	if filepath.IsAbs(resource.Path) {
		return fmt.Errorf("resources must use a http or https URL or a relative path: absolute path not supported: %s", resource)
	}

	u, err := url.Parse(resource.Path)
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
