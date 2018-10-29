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
	"net/url"
	"path/filepath"

	"github.com/projectriff/riff/pkg/fileutils"

	"github.com/ghodss/yaml"
)

const manifestVersion_0_1 = "0.1"

var manifests = map[string]*Manifest{
	"latest": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/knative-releases/serving/latest/istio.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/build/latest/release.yaml",
			"https://storage.googleapis.com/knative-releases/serving/latest/serving.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/latest/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/latest/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/latest/riff-build.yaml",
			"https://storage.googleapis.com/riff-releases/riff-cnb-buildtemplate-0.1.0.pre.3.yaml",
		},
	},
	"stable": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20181029-1d5c521/istio.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/build/previous/v20181029-e9f5b24/release.yaml",
			"https://storage.googleapis.com/knative-releases/serving/previous/v20181029-1d5c521/serving.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20181029-642bfc1/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20181029-642bfc1/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
			"https://storage.googleapis.com/riff-releases/riff-cnb-buildtemplate-0.1.0.yaml",
		},
	},
	"v0.1.3": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180921-69811e7/istio.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180921-69811e7/release-no-mon.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180921-01f95cb/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180921-01f95cb/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
			"https://storage.googleapis.com/riff-releases/riff-cnb-buildtemplate-0.1.0.yaml",
		},
	},
	"v0.1.2": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/istio.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/release-no-mon.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180830-5d35af5/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180830-5d35af5/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
		},
	},
	"v0.1.1": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/riff-releases/istio/istio-1.0.0-riff-crds.yaml",
			"https://storage.googleapis.com/riff-releases/istio/istio-1.0.0-riff-main.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/knative-releases/serving/previous/v20180809-6b01d8e/release-no-mon.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180809-34ab480/release.yaml",
			"https://storage.googleapis.com/knative-releases/eventing/previous/v20180809-34ab480/release-clusterbus-stub.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
		},
	},
	"v0.1.0": &Manifest{
		ManifestVersion: manifestVersion_0_1,
		Istio: []string{
			"https://storage.googleapis.com/riff-releases/istio-riff-0.1.0.yaml",
		},
		Knative: []string{
			"https://storage.googleapis.com/riff-releases/release-no-mon-riff-0.1.0.yaml",
			"https://storage.googleapis.com/riff-releases/release-eventing-riff-0.1.0.yaml",
			"https://storage.googleapis.com/riff-releases/release-eventing-clusterbus-stub-riff-0.1.0.yaml",
		},
		Namespace: []string{
			"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
		},
	},
}

// Manifest defines the location of YAML files for system components.
type Manifest struct {
	ManifestVersion string   `json:"manifestVersion"`
	Istio           []string `json:"istio"`
	Knative         []string `json:"knative"`
	Namespace       []string `json:"namespace"`
}

func NewManifest(path string) (*Manifest, error) {
	if manifest, ok := manifests[path]; ok {
		return manifest, nil
	}

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
	return fmt.Errorf("Manifest is incomplete: %s array missing: %#v", omission, m)
}

func checkResource(resource string) error {
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
