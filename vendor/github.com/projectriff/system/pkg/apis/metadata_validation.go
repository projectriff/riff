/*
Copyright 2018 The Knative Authors

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

package apis

import (
	"strings"

	"github.com/knative/pkg/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	maxLength = 63
)

func ValidateObjectMetadata(meta metav1.Object) *apis.FieldError {
	name := meta.GetName()
	generateName := meta.GetGenerateName()
	errs := &apis.FieldError{}

	if name == "" && generateName == "" {
		errs = errs.Also(apis.ErrMissingOneOf("name", "generateName"))
	}

	if strings.Contains(name, ".") {
		errs = errs.Also(apis.ErrInvalidValue("special character . must not be present", "name"))
	}
	if strings.Contains(generateName, ".") {
		errs = errs.Also(apis.ErrInvalidValue("special character . must not be present", "generateName"))
	}

	if len(name) > maxLength {
		errs = errs.Also(apis.ErrInvalidValue("length must be no more than 63 characters", "name"))
	}
	if len(generateName) > maxLength {
		errs = errs.Also(apis.ErrInvalidValue("length must be no more than 63 characters", "generateName"))
	}

	return errs
}
