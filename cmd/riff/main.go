/*
 * Copyright 2018-2019 The original author or authors
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

package main

import (
	"fmt"
	"os"

	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/riff/commands"
)

var (
	manifests = map[string]*core.Manifest{
		// validated, compatible versions of Knative
		"stable": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v0.4.0/istio.yaml",
				"https://storage.googleapis.com/projectriff/istio/istio-riff-knative-serving-v0-4-0-patch.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/previous/v0.4.0/build.yaml",
				"https://storage.googleapis.com/knative-releases/serving/previous/v0.4.1/serving.yaml",
				"https://raw.githubusercontent.com/knative/serving/v0.4.0/third_party/config/build/clusterrole.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v0.4.0/eventing.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v0.4.0/in-memory-channel.yaml",
				// TODO update to a release version before releasing riff
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-clusterbuildtemplate-0.2.0-snapshot-ci-92b29a6fb99c.yaml",
			},
		},
		// most recent release of Knative. This manifest is not tested
		"latest": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/latest/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/latest/build.yaml",
				"https://storage.googleapis.com/knative-releases/serving/latest/serving.yaml",
				"https://raw.githubusercontent.com/knative/serving/master/third_party/config/build/clusterrole.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/latest/eventing.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/latest/in-memory-channel.yaml",
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-clusterbuildtemplate.yaml",
			},
		},
		// most recent build of Knative from master. This manifest is not tested
		"nightly": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-nightly/serving/latest/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-nightly/build/latest/build.yaml",
				"https://storage.googleapis.com/knative-nightly/serving/latest/serving.yaml",
				"https://raw.githubusercontent.com/knative/serving/master/third_party/config/build/clusterrole.yaml",
				"https://storage.googleapis.com/knative-nightly/eventing/latest/eventing.yaml",
				"https://storage.googleapis.com/knative-nightly/eventing/latest/in-memory-channel.yaml",
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-clusterbuildtemplate.yaml",
			},
		},
	}
)

func main() {

	root := commands.CreateAndWireRootCommand(manifests)

	sub, err := root.ExecuteC()
	if err != nil {
		if !sub.SilenceUsage { // May have been switched to true once we're past PreRunE()
			sub.Help()
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}
