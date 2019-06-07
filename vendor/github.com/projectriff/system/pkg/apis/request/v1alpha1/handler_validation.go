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
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/knative/pkg/apis"
	systemapis "github.com/projectriff/system/pkg/apis"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func (h *Handler) Validate(ctx context.Context) *apis.FieldError {
	errs := &apis.FieldError{}
	errs = errs.Also(systemapis.ValidateObjectMetadata(h.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(h.Spec.Validate(ctx).ViaField("spec"))
	return errs
}

func (hs HandlerSpec) Validate(ctx context.Context) *apis.FieldError {
	if equality.Semantic.DeepEqual(hs, HandlerSpec{}) {
		return apis.ErrMissingField(apis.CurrentField)
	}

	errs := &apis.FieldError{}

	if diff := cmp.Diff(&corev1.PodSpec{
		// add supported PodSpec fields here, otherwise their usage will be rejected
		ServiceAccountName: hs.Template.ServiceAccountName,
		// the defaulter guarantees at least one container
		Containers: filterInvalidContainers(hs.Template.Containers[:1]),
		Volumes:    filterInvalidVolumes(hs.Template.Volumes),
	}, hs.Template); diff != "" {
		err := apis.ErrDisallowedFields(apis.CurrentField)
		err.Details = fmt.Sprintf("limited Template fields may be set (-want, +got) = %v", diff)
		errs = errs.Also(err)
	}

	if hs.Build == nil && hs.Template.Containers[0].Image == "" {
		errs = errs.Also(apis.ErrMissingOneOf("build", "template.containers[0].image"))
	} else if hs.Build != nil && hs.Template.Containers[0].Image != "" {
		errs = errs.Also(apis.ErrMultipleOneOf("build", "template.containers[0].image"))
	} else if hs.Build != nil {
		errs = errs.Also(hs.Build.Validate(ctx).ViaField("build"))
	}

	return errs
}

func (b *Build) Validate(ctx context.Context) *apis.FieldError {
	if equality.Semantic.DeepEqual(b, &Build{}) {
		return apis.ErrMissingField(apis.CurrentField)
	}

	errs := &apis.FieldError{}
	used := []string{}
	unused := []string{}

	if b.ApplicationRef != "" {
		used = append(used, "applicationRef")
	} else {
		unused = append(unused, "applicationRef")
	}

	if b.FunctionRef != "" {
		used = append(used, "functionRef")
	} else {
		unused = append(unused, "functionRef")
	}

	if len(used) == 0 {
		errs = errs.Also(apis.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(apis.ErrMultipleOneOf(used...))
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
