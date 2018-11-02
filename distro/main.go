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
	"os"

	"fmt"

	"github.com/projectriff/riff/distro/commands"
	"github.com/projectriff/riff/pkg/core"
)

var (
	manifests = map[string]*core.Manifest{
		"latest": &core.Manifest{
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
				"https://storage.googleapis.com/riff-releases/riff-cnb-buildtemplate-0.1.0.yaml",
			},
		},
	}
)

func main() {

	root := commands.DistroCreateAndWireRootCommand(manifests)

	err := root.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
