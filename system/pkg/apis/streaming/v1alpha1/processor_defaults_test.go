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

func TestProcessorDefault(t *testing.T) {
	tests := []struct {
		name string
		in   *Processor
		want *Processor
	}{{
		name: "empty",
		in:   &Processor{},
		want: &Processor{
			Spec: ProcessorSpec{
				Inputs:  []InputStreamBinding{},
				Outputs: []OutputStreamBinding{},
				Template: &corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
						Labels:      map[string]string{},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "function"},
						},
						Volumes: []corev1.Volume{},
					},
				},
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

func TestProcessorSpecDefault(t *testing.T) {
	tests := []struct {
		name string
		in   *ProcessorSpec
		want *ProcessorSpec
	}{{
		name: "empty",
		in:   &ProcessorSpec{},
		want: &ProcessorSpec{
			Inputs:  []InputStreamBinding{},
			Outputs: []OutputStreamBinding{},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
					Volumes: []corev1.Volume{},
				},
			},
		},
	}, {
		name: "add alias",
		in: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-input"},
			},
			Outputs: []OutputStreamBinding{
				{Stream: "my-output"},
			}},
		want: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-input", Alias: "my-input", StartOffset: Latest},
			},
			Outputs: []OutputStreamBinding{
				{Stream: "my-output", Alias: "my-output"},
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
					Volumes: []corev1.Volume{},
				},
			},
		},
	}, {
		name: "preserves alias",
		in: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-input", Alias: "in"},
			},
			Outputs: []OutputStreamBinding{
				{Stream: "my-output", Alias: "out"},
			}},
		want: &ProcessorSpec{
			Inputs: []InputStreamBinding{
				{Stream: "my-input", Alias: "in", StartOffset: Latest},
			},
			Outputs: []OutputStreamBinding{
				{Stream: "my-output", Alias: "out"},
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
					Volumes: []corev1.Volume{},
				},
			},
		},
	}, {
		name: "add container name",
		in: &ProcessorSpec{
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
			},
		},
		want: &ProcessorSpec{
			Inputs:  []InputStreamBinding{},
			Outputs: []OutputStreamBinding{},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "function"},
					},
					Volumes: []corev1.Volume{},
				},
			},
		},
	}, {
		name: "preserves container",
		in: &ProcessorSpec{
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "function",
							Env: []corev1.EnvVar{
								{Name: "MY_VAR", Value: "my-value"},
							},
						},
					},
				},
			},
		},
		want: &ProcessorSpec{
			Inputs:  []InputStreamBinding{},
			Outputs: []OutputStreamBinding{},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
					Labels:      map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "function",
							Env: []corev1.EnvVar{
								{Name: "MY_VAR", Value: "my-value"},
							},
						},
					},
					Volumes: []corev1.Volume{},
				},
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
