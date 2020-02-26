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
	"fmt"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/projectriff/system/pkg/validation"
)

// +kubebuilder:webhook:path=/validate-core-projectriff-io-v1alpha1-deployer,mutating=false,failurePolicy=fail,groups=core.projectriff.io,resources=deployers,verbs=create;update,versions=v1alpha1,name=deployers.core.projectriff.io

var (
	_ webhook.Validator         = &Deployer{}
	_ validation.FieldValidator = &Deployer{}
)

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Deployer) ValidateCreate() error {
	return r.Validate().ToAggregate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Deployer) ValidateUpdate(old runtime.Object) error {
	// TODO check for immutable fields
	return r.Validate().ToAggregate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Deployer) ValidateDelete() error {
	return nil
}

func (r *Deployer) Validate() validation.FieldErrors {
	errs := validation.FieldErrors{}

	errs = errs.Also(r.Spec.Validate().ViaField("spec"))

	return errs
}

func (s *DeployerSpec) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(s, &DeployerSpec{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}

	if diff := cmp.Diff(&corev1.PodSpec{
		// add supported PodSpec fields here, otherwise their usage will be rejected
		ServiceAccountName: s.Template.Spec.ServiceAccountName,
		// the defaulter guarantees at least one container
		Containers: filterInvalidContainers(s.Template.Spec.Containers[:1]),
		Volumes:    filterInvalidVolumes(s.Template.Spec.Volumes),
	}, &s.Template.Spec); diff != "" {
		errs = errs.Also(validation.ErrDisallowedFields("template.spec", fmt.Sprintf("limited PodSpec fields may be set (-want, +got) = %v", diff)))
	}

	for i, env := range s.Template.Spec.Containers[0].Env {
		if env.Name == "PORT" {
			errs = errs.Also(validation.ErrDisallowedFields(fmt.Sprintf("template.spec.containers[0].env[%d]", i), "PORT is not allowed"))
		}
	}

	if s.Build == nil && s.Template.Spec.Containers[0].Image == "" {
		errs = errs.Also(validation.ErrMissingOneOf("build", "template.spec.containers[0].image"))
	} else if s.Build != nil && s.Template.Spec.Containers[0].Image != "" {
		errs = errs.Also(validation.ErrMultipleOneOf("build", "template.spec.containers[0].image"))
	} else if s.Build != nil {
		errs = errs.Also(s.Build.Validate().ViaField("build"))
	}

	if s.IngressPolicy != "" && s.IngressPolicy != IngressPolicyClusterLocal && s.IngressPolicy != IngressPolicyExternal {
		errs = errs.Also(validation.ErrInvalidValue(s.IngressPolicy, "ingressPolicy"))
	}

	return errs
}

func (b *Build) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(b, &Build{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}
	used := []string{}
	unused := []string{}

	if b.ApplicationRef != "" {
		used = append(used, "applicationRef")
	} else {
		unused = append(unused, "applicationRef")
	}

	if b.ContainerRef != "" {
		used = append(used, "containerRef")
	} else {
		unused = append(unused, "containerRef")
	}

	if b.FunctionRef != "" {
		used = append(used, "functionRef")
	} else {
		unused = append(unused, "functionRef")
	}

	if len(used) == 0 {
		errs = errs.Also(validation.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(validation.ErrMultipleOneOf(used...))
	}

	return errs
}

func filterInvalidContainers(containers []corev1.Container) []corev1.Container {
	// TODO remove unsupported fields
	return containers
}

func filterInvalidVolumes(volumes []corev1.Volume) []corev1.Volume {
	// TODO remove unsupported fields
	return volumes
}
