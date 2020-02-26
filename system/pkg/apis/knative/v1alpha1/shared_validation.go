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

	"github.com/projectriff/riff/system/pkg/validation"
)

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
