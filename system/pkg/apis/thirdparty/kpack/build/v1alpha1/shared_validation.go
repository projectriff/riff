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

	"github.com/projectriff/system/pkg/validation"
)

func (s *SourceConfig) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(s, &SourceConfig{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}
	used := []string{}
	unused := []string{}

	if s.Git != nil {
		used = append(used, "git")
		errs = errs.Also(s.Git.Validate().ViaField("git"))
	} else {
		unused = append(unused, "git")
	}

	if s.Blob != nil {
		used = append(used, "git")
		errs = errs.Also(s.Blob.Validate().ViaField("blob"))
	} else {
		unused = append(unused, "blob")
	}

	if s.Registry != nil {
		used = append(used, "git")
		errs = errs.Also(s.Registry.Validate().ViaField("registry"))
	} else {
		unused = append(unused, "registry")
	}

	if len(used) == 0 {
		errs = errs.Also(validation.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(validation.ErrMultipleOneOf(used...))
	}

	return errs
}

func (g *Git) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(g, &Git{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}

	if g.URL == "" {
		errs = errs.Also(validation.ErrMissingField("url"))
	}

	if g.Revision == "" {
		errs = errs.Also(validation.ErrMissingField("revision"))
	}

	return errs
}

func (b *Blob) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(b, &Blob{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}

	if b.URL == "" {
		errs = errs.Also(validation.ErrMissingField("url"))
	}
	// TODO add more validation as to the type of URL

	return errs
}

func (r *Registry) Validate() validation.FieldErrors {
	if equality.Semantic.DeepEqual(r, &Registry{}) {
		return validation.ErrMissingField(validation.CurrentField)
	}

	errs := validation.FieldErrors{}

	if r.Image == "" {
		errs = errs.Also(validation.ErrMissingField("image"))
	}

	for i, s := range r.ImagePullSecrets {
		if s.Name == "" {
			errs = errs.Also(validation.ErrMissingField("name").ViaIndex(i))
		}
	}

	return errs
}
