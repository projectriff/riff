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
	"fmt"
	"github.com/projectriff/riff/pkg/constants"
	"os"

	"github.com/projectriff/riff/pkg/core"

	"github.com/projectriff/riff/cmd/commands"
)

var (
	Manifests = map[string]*core.Manifest{
		// validated, compatible versions of Knative
		"stable": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v0.2.3/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/previous/v0.2.0/release.yaml",
				"https://storage.googleapis.com/knative-releases/serving/previous/v0.2.3/serving.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v0.2.1/eventing.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v0.2.1/in-memory-channel.yaml",
				fmt.Sprintf("https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-clusterbuildtemplate-%s.yaml", constants.BuilderVersion),
			},
			Namespace: []string{
				fmt.Sprintf("https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-cache-%s.yaml", constants.BuilderVersion),
			},
		},
		// most recent release of Knative. This manifest is not tested
		"latest": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/latest/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/latest/release.yaml",
				"https://storage.googleapis.com/knative-releases/serving/latest/serving.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/latest/eventing.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/latest/in-memory-channel.yaml",
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-clusterbuildtemplate.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-cache.yaml",
			},
		},
		// most recent build of Knative from master. This manifest is not tested
		"nightly": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-nightly/serving/latest/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-nightly/build/latest/release.yaml",
				"https://storage.googleapis.com/knative-nightly/serving/latest/serving.yaml",
				"https://storage.googleapis.com/knative-nightly/eventing/latest/eventing.yaml",
				"https://storage.googleapis.com/knative-nightly/eventing/latest/in-memory-channel.yaml",
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-clusterbuildtemplate.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-cache.yaml",
			},
		},
	}
)

func main() {

	root := commands.CreateAndWireRootCommand(Manifests, constants.BuilderName, constants.DefaultRunImage)

	sub, err := root.ExecuteC()
	if err != nil {
		if !sub.SilenceUsage { // May have been switched to true once we're past PreRunE()
			sub.Help()
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}
