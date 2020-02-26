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

	"github.com/projectriff/riff/system/pkg/validation"
)

// +kubebuilder:webhook:path=/validate-build-projectriff-io-v1alpha1-application,mutating=false,failurePolicy=fail,groups=build.projectriff.io,resources=applications,verbs=create;update,versions=v1alpha1,name=applications.build.projectriff.io

var (
	_ webhook.Validator         = &Application{}
	_ validation.FieldValidator = &Application{}
)

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateCreate() error {
	return r.Validate().ToAggregate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateUpdate(old runtime.Object) error {
	// TODO check for immutable fields
	return r.Validate().ToAggregate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateDelete() error {
	return nil
}

func (r *Application) Validate() validation.FieldErrors {
	errs := validation.FieldErrors{}

	errs = errs.Also(r.Spec.Validate().ViaField("spec"))

	return errs
}

func (s *ApplicationSpec) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(s, &ApplicationSpec{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}

	if s.Image == "" {
		errs = errs.Also(validation.ErrMissingField("image"))
	}

	if s.Source != nil {
		errs = errs.Also(s.Source.Validate().ViaField("source"))
	}

	return errs
}
