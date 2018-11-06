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

package main

import (
	"github.com/projectriff/riff/pkg/core"
)

/*

The following declarations determine the docker images and kubernetes configuration files that are used by all packages
in this repository. To preserve this consistency, the declarations are shared between the main packages. This ensures
that other packages do not import the declarations directly but must rely on the values being injected and passed as
parameters.

This file is defined in the root of this repository and symbolically linked from other main packages such as that in
./distro.

*/

var (
	buildpacks = map[string]string{
		"java":    "projectriff/buildpack",
		"command": "projectriff/buildpack",
		"node":    "projectriff/buildpack",
	}
	invokers = map[string]string{
		"jar": "https://github.com/projectriff/java-function-invoker/raw/v0.1.1/java-invoker.yaml",
	}
	manifests = map[string]*core.Manifest{
		"latest": {
			ManifestVersion: "0.1",
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
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-buildtemplate-0.0.1-snapshot-ci-64830c3bbc6503beafdae382ead115806fa100ca.yaml",
			},
		},
		"stable": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v0.2.0/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/previous/v0.2.0/release.yaml",
				"https://storage.googleapis.com/knative-releases/serving/previous/v20181101-v0.2.0-11-g877523d/serving.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20181031-a2f9417/release.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20181031-a2f9417/release-clusterbus-stub.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/riff-releases/previous/riff-build/riff-build-0.1.0.yaml",
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-buildtemplate-0.0.1-snapshot-ci-64830c3bbc6503beafdae382ead115806fa100ca.yaml",
			},
		},
	}
)
