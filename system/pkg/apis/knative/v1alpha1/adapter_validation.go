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
	"k8s.io/apimachinery/pkg/api/equality"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/projectriff/system/pkg/validation"
)

// +kubebuilder:webhook:path=/validate-knative-projectriff-io-v1alpha1-adapter,mutating=false,failurePolicy=fail,groups=knative.projectriff.io,resources=adapters,verbs=create;update,versions=v1alpha1,name=adapters.knative.projectriff.io

var (
	_ webhook.Validator         = &Adapter{}
	_ validation.FieldValidator = &Adapter{}
)

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Adapter) ValidateCreate() error {
	return r.Validate().ToAggregate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Adapter) ValidateUpdate(old runtime.Object) error {
	// TODO check for immutable fields
	return r.Validate().ToAggregate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Adapter) ValidateDelete() error {
	return nil
}

func (r *Adapter) Validate() validation.FieldErrors {
	errs := validation.FieldErrors{}

	errs = errs.Also(r.Spec.Validate().ViaField("spec"))

	return errs
}

func (s AdapterSpec) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(s, AdapterSpec{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}

	errs = errs.Also(s.Build.Validate().ViaField("build"))
	errs = errs.Also(s.Target.Validate().ViaField("target"))

	return errs
}

func (t *AdapterTarget) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(t, &AdapterTarget{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}
	used := []string{}
	unused := []string{}

	if t.ServiceRef != "" {
		used = append(used, "serviceRef")
	} else {
		unused = append(unused, "serviceRef")
	}

	if t.ConfigurationRef != "" {
		used = append(used, "configurationRef")
	} else {
		unused = append(unused, "configurationRef")
	}

	if len(used) == 0 {
		errs = errs.Also(validation.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(validation.ErrMultipleOneOf(used...))
	}

	return errs
}
