/*
Copyright 2019 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeployerDefault(t *testing.T) {
	tests := []struct {
		name string
		in   *Deployer
		want *Deployer
	}{{
		name: "empty",
		in:   &Deployer{},
		want: &Deployer{
			Spec: DeployerSpec{
				Template: &corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
						Labels:      map[string]string{},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{},
						},
					},
				},
				IngressPolicy: IngressPolicyClusterLocal,
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.in
			got.Default()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Default (-want, +got) = %v", diff)
			}
		})
	}
}

func TestDeployerSpecDefault(t *testing.T) {
	tests := []struct {
		name string
		in   *DeployerSpec
		want *DeployerSpec
	}{{
		name: "ensure at least one container",
		in:   &DeployerSpec{},
		want: &DeployerSpec{
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
			},
			IngressPolicy: IngressPolicyClusterLocal,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.in
			got.Default()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Default (-want, +got) = %v", diff)
			}
		})
	}
}
