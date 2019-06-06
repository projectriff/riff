/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	"context"

	"github.com/knative/pkg/apis"
	"k8s.io/apimachinery/pkg/api/equality"
)

func (ba *BuildArgument) Validate(ctx context.Context) *apis.FieldError {
	if equality.Semantic.DeepEqual(ba, &BuildArgument{}) {
		return apis.ErrMissingField(apis.CurrentField)
	}

	errs := &apis.FieldError{}
	if ba.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	}
	return errs
}

func (s *Source) Validate(ctx context.Context) *apis.FieldError {
	if equality.Semantic.DeepEqual(s, &Source{}) {
		return apis.ErrMissingField(apis.CurrentField)
	}

	errs := &apis.FieldError{}
	used := []string{}
	unused := []string{}

	if s.Git != nil {
		used = append(used, "git")
		errs = errs.Also(s.Git.Validate(ctx).ViaField("git"))
	} else {
		unused = append(unused, "git")
	}

	if len(used) == 0 {
		errs = errs.Also(apis.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(apis.ErrMultipleOneOf(used...))
	}

	return errs
}

func (gs *GitSource) Validate(ctx context.Context) *apis.FieldError {
	if equality.Semantic.DeepEqual(gs, &GitSource{}) {
		return apis.ErrMissingField(apis.CurrentField)
	}

	errs := &apis.FieldError{}

	if gs.URL == "" {
		errs = errs.Also(apis.ErrMissingField("url"))
	}

	if gs.Revision == "" {
		errs = errs.Also(apis.ErrMissingField("revision"))
	}

	return errs
}
