/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import (
	"github.com/knative/pkg/apis"
	"k8s.io/apimachinery/pkg/api/validation"
)

func K8sName(name, field string) *apis.FieldError {
	errs := &apis.FieldError{}

	if out := validation.NameIsDNSLabel(name, false); len(out) != 0 {
		// TODO capture info about why the name is invalid
		errs = errs.Also(apis.ErrInvalidValue(name, field))
	}

	return errs
}

func K8sNames(names []string, field string) *apis.FieldError {
	errs := &apis.FieldError{}

	for i, name := range names {
		if name == "" {
			errs = errs.Also(apis.ErrInvalidValue(name, apis.CurrentField).ViaFieldIndex(field, i))
		} else {
			errs = errs.Also(K8sName(name, apis.CurrentField).ViaFieldIndex(field, i))
		}
	}

	return errs
}
