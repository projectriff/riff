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
	systemapis "github.com/projectriff/system/pkg/apis"
	"k8s.io/apimachinery/pkg/api/equality"
)

func (p *Processor) Validate(ctx context.Context) *apis.FieldError {
	errs := &apis.FieldError{}
	errs = errs.Also(systemapis.ValidateObjectMetadata(p.GetObjectMeta()).ViaField("metadata"))
	errs = errs.Also(p.Spec.Validate(ctx).ViaField("spec"))
	return errs
}

func (ps *ProcessorSpec) Validate(ctx context.Context) *apis.FieldError {
	if equality.Semantic.DeepEqual(ps, &ProcessorSpec{}) {
		return apis.ErrMissingField(apis.CurrentField)
	}

	errs := &apis.FieldError{}

	if ps.FunctionRef == "" {
		errs = errs.Also(apis.ErrMissingField("functionRef"))
	}

	// at least one input is required
	if len(ps.Inputs) == 0 {
		errs = errs.Also(apis.ErrMissingField("inputs"))
	}
	for i, input := range ps.Inputs {
		if input == "" {
			errs = errs.Also(apis.ErrInvalidArrayValue(input, "inputs", i))
		}
	}

	// outputs are optional
	for i, output := range ps.Outputs {
		if output == "" {
			errs = errs.Also(apis.ErrInvalidArrayValue(output, "outputs", i))
		}
	}

	return errs
}
