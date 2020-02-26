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

	"github.com/projectriff/system/pkg/validation"
)

func TestValidateDeployer(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *Deployer
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &Deployer{},
		expected: validation.ErrMissingField("spec"),
	}, {
		name: "valid",
		target: &Deployer{
			Spec: DeployerSpec{
				Build: &Build{
					FunctionRef: "my-function",
				},
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{},
						},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateDeployer(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}

func TestValidateDeployerSpec(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *DeployerSpec
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &DeployerSpec{},
		expected: validation.ErrMissingField(validation.CurrentField),
	}, {
		name: "valid, build",
		target: &DeployerSpec{
			Build: &Build{
				FunctionRef: "my-function",
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "valid, container image",
		target: &DeployerSpec{
			Build: nil,
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "my-iamge"},
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "invalid, build and container image",
		target: &DeployerSpec{
			Build: &Build{
				FunctionRef: "my-function",
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "my-image"},
					},
				},
			},
		},
		expected: validation.ErrMultipleOneOf("build", "template.spec.containers[0].image"),
	}, {
		name: "no PORT env",
		target: &DeployerSpec{
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "my-iamge",
							Env: []corev1.EnvVar{
								{Name: "PORT", Value: "8080"},
							},
						},
					},
				},
			},
		},
		expected: validation.ErrDisallowedFields("template.spec.containers[0].env[0]", "PORT is not allowed"),
	}, {
		name: "invalid, ingress policy",
		target: &DeployerSpec{
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "my-iamge"},
					},
				},
			},
			IngressPolicy: "bogus",
		},
		expected: validation.ErrInvalidValue(IngressPolicy("bogus"), "ingressPolicy"),
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateDeployerSpec(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}
