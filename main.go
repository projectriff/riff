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
	"github.com/projectriff/riff/pkg/crd"
	"github.com/projectriff/riff/pkg/env"
	"os"

	"github.com/projectriff/riff/cmd/commands"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// TODO update to a release version before releasing riff
	builderVersion  = "0.2.0-snapshot-ci-63cd05079e1f"
	builderName         = fmt.Sprintf("projectriff/builder:%s", builderVersion)
	defaultRunImage = "packs/run:v3alpha2"

	manifests = map[string]*crd.Manifest {
		"stable": {
			ObjectMeta: metav1.ObjectMeta{
				Name:   env.Cli.Name + "-install",
				Labels: map[string]string{env.Cli.Name + "-install": "true"},
			},
			TypeMeta: metav1.TypeMeta{
				Kind: crd.Kind,
				APIVersion: fmt.Sprintf("%s/%s", crd.Group, crd.Version),
			},
			Spec: crd.RiffSpec {
				Resources: []crd.RiffResource{
					{
						Path: "https://storage.googleapis.com/knative-releases/serving/previous/v0.3.0/istio.yaml",
						Name: "istio",
						Namespace: "istio-system",
						Checks: []crd.ResourceChecks {
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"istio": "citadel"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"istio": "egressgateway"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"istio": "galley"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"istio": "ingressgateway"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"istio": "pilot"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"istio-mixer-type": "policy"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"istio": "sidecar-injector"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"istio-mixer-type": "telemetry"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
						},
					},
					{
						Path: "https://storage.googleapis.com/projectriff/istio/istio-riff-knative-serving-v0-3-0-patch.yaml",
						Name: "istio-riff-patch",
					},
					{
						// NOTE: build should be in the knative-releases bucket, but is hiding in knative-nightly
						Path: "https://storage.googleapis.com/knative-nightly/build/previous/v0.3.0/release.yaml",
						Name: "build",
						Namespace: "knative-build",
						Checks: []crd.ResourceChecks {
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "build-controller"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "build-webhook"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
						},
					},
					{
						Path: "https://storage.googleapis.com/knative-releases/serving/previous/v0.3.0/serving.yaml",
						Name: "serving",
						Namespace: "knative-serving",
						Checks: []crd.ResourceChecks {
							{
								Kind:      "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "activator"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind:      "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "autoscaler"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "controller"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "webhook"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
						},
					},
					{
						Path: "https://storage.googleapis.com/knative-releases/eventing/previous/v0.3.0/eventing.yaml",
						Name: "eventing",
						Namespace: "knative-eventing",
						Checks: []crd.ResourceChecks {
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "eventing-controller"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "webhook"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
						},
					},
					{
						Path: "https://storage.googleapis.com/knative-releases/eventing/previous/v0.3.0/in-memory-channel.yaml",
						Name: "eventing-in-memory-channel",
						Namespace: "knative-eventing",
						Checks: []crd.ResourceChecks{
							{
								Kind:      "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"role": "dispatcher", "clusterChannelProvisioner": "in-memory-channel"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
							{
								Kind: "Pod",
								Selector: metav1.LabelSelector{
									MatchLabels: map[string]string{"role": "controller", "clusterChannelProvisioner":"in-memory-channel"},
								},
								JsonPath: ".status.phase",
								Pattern:  "Running",
							},
						},
					},
					{
						Path: fmt.Sprintf("https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-clusterbuildtemplate-%s.yaml", builderVersion),
						Name: "riff-build-template",
					},
				},
				Init: []crd.RiffResource {
					{
						Path: fmt.Sprintf("https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-cache-%s.yaml", builderVersion),
						Name: "riff-build-cache",
					},
				},
			},
		},
	}
)

func main() {

	root := commands.CreateAndWireRootCommand(manifests, builderName, defaultRunImage)

	sub, err := root.ExecuteC()
	if err != nil {
		if !sub.SilenceUsage { // May have been switched to true once we're past PreRunE()
			sub.Help()
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}
}
