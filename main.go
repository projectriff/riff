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
	"os"

	"github.com/projectriff/riff/pkg/core"

	"github.com/projectriff/riff/cmd/commands"
)

var (
	builder = "projectriff/builder"

	manifests = map[string]*core.Manifest{
		"latest": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/latest/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/latest/release.yaml",
				"https://storage.googleapis.com/knative-releases/serving/latest/serving.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20181106-a99376f/release.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20181106-a99376f/release-clusterbus-stub.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-buildtemplate.yaml",
			},
		},
		"stable": {
			ManifestVersion: "0.1",
			Istio: []string{
				"https://storage.googleapis.com/knative-releases/serving/previous/v20181106-v0.2.0-26-g91cd00d/istio.yaml",
			},
			Knative: []string{
				"https://storage.googleapis.com/knative-releases/build/previous/v0.2.0/release.yaml",
				"https://storage.googleapis.com/knative-releases/serving/previous/v20181106-v0.2.0-26-g91cd00d/serving.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20181031-a2f9417/release.yaml",
				"https://storage.googleapis.com/knative-releases/eventing/previous/v20181031-a2f9417/release-clusterbus-stub.yaml",
			},
			Namespace: []string{
				"https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-buildtemplate-0.0.1-snapshot-ci-04852871adf05191969bb3bee8ed65cf7cd31285.yaml",
			},
		},
	}
)

func main() {

	root := commands.CreateAndWireRootCommand(manifests, builder)

	sub, err := root.ExecuteC()
	if err != nil {
		if !sub.SilenceUsage { // May have been switched to true once we're past PreRunE()
			sub.Help()
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}
