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

package crd

import (
	"github.com/jinzhu/copier"
	"github.com/projectriff/riff/pkg/env"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceChecks struct {
	Kind       string               `json:"kind,omitempty"`
	Namespace  string               `json:"namespace,omitempty"`
	Selector   metav1.LabelSelector `json:"selector,omitempty"`
	JsonPath   string               `json:"jsonpath,omitempty"`
	Pattern    string               `json:"pattern,omitempty"`
}

type RiffResources struct {
	Path       string           `json:"path,omitempty"`
	Name       string           `json:"name,omitempty"`
	Checks     []ResourceChecks `json:"checks,omitempty"`
}

type RiffSpec struct {
	Images    []string        `json:"images,omitempty"`
	Resources []RiffResources `json:"resources,omitempty"`
}

type RiffStatus struct {
	Status string `json:"status,omitempty"`
}

type RiffManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RiffSpec   `json:"spec,omitempty"`
	Status RiffStatus `json:"status,omitempty"`
}

func (orig *RiffManifest) DeepCopyObject() runtime.Object {
	result := &RiffManifest{}
	copier.Copy(result, orig)
	return result
}

var schemeGroupVersion = schema.GroupVersion{
	Group:    Group,
	Version:  Version,
}

func NewManifest() *RiffManifest {
	manifest := &RiffManifest{
		ObjectMeta: metav1.ObjectMeta{
			Name:   env.Cli.Name + "-install",
			Labels: map[string]string{env.Cli.Name + "-install": "true"},
		},
		TypeMeta: metav1.TypeMeta{
			Kind: Kind,
			APIVersion: "projectriff.io/" + Version,
		},
		Spec: RiffSpec{
			Resources: []RiffResources{
				{
					Path: "https://storage.googleapis.com/knative-releases/serving/previous/v0.2.2/istio.yaml",
					Name: "istio",
					Checks: []ResourceChecks{
						{
							Kind: "Pod",
							Namespace: "istio-system",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"istio": "citadel"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "istio-system",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"istio": "egressgateway"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "istio-system",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"istio": "galley"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "istio-system",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"istio": "ingressgateway"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "istio-system",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"istio": "pilot"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "istio-system",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"istio-mixer-type": "policy"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "istio-system",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"istio": "sidecar-injector"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "istio-system",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"istio-mixer-type": "telemetry"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
					},
				},
				{
					Path: "https://storage.googleapis.com/knative-releases/build/latest/release.yaml",
					Name: "build",
					Checks: []ResourceChecks{
						{
							Kind: "Pod",
							Namespace: "knative-build",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "build-controller"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "knative-build",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "build-webhook"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
					},
				},
				{
					Path: "https://storage.googleapis.com/knative-releases/serving/latest/serving.yaml",
					Name: "serving",
					Checks: []ResourceChecks{
						{
							Kind: "Pod",
							Namespace: "knative-serving",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "activator"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "knative-serving",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "autoscaler"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "knative-serving",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "controller"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "knative-serving",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "webhook"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
					},
				},
				{
					Path: "https://storage.googleapis.com/knative-releases/eventing/latest/eventing.yaml",
					Name: "eventing",
					Checks: []ResourceChecks{
						{
							Kind: "Pod",
							Namespace: "knative-eventing",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "eventing-controller"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "knative-eventing",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "webhook"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
					},
				},
				{
					Path: "https://storage.googleapis.com/knative-releases/eventing/latest/in-memory-channel.yaml",
					Name: "eventing-in-memory-channel",
					Checks: []ResourceChecks{
						{
							Kind: "Pod",
							Namespace: "knative-eventing",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"role": "dispatcher", "clusterChannelProvisioner":"in-memory-channel"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
						{
							Kind: "Pod",
							Namespace: "knative-eventing",
							Selector: metav1.LabelSelector{
								MatchLabels: map[string]string{"role": "controller", "clusterChannelProvisioner":"in-memory-channel"},
							},
							JsonPath: ".status.phase",
							Pattern:  "Running",
						},
					},
				},
				{
					Path: "https://storage.googleapis.com/projectriff/riff-buildtemplate/riff-cnb-buildtemplate.yaml",
					Name: "riff-build-template",
				},
			},
		},
	}
	return manifest
}
