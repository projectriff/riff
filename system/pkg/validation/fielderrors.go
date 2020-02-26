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

package validation

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const CurrentField = ""

type FieldValidator interface {
	Validate() FieldErrors
}

type FieldErrors field.ErrorList

func (e FieldErrors) Also(errs ...FieldErrors) FieldErrors {
	aggregate := e
	for _, err := range errs {
		aggregate = append(aggregate, err...)
	}
	return aggregate
}

func (e FieldErrors) ViaField(key string) FieldErrors {
	errs := make(FieldErrors, len(e))
	for i, err := range e {
		newField := key
		if !strings.HasPrefix(err.Field, "[") {
			newField = newField + "."
		}
		if err.Field != "[]" {
			newField = newField + err.Field
		}
		errs[i] = &field.Error{
			Type:     err.Type,
			Field:    newField,
			BadValue: err.BadValue,
			Detail:   err.Detail,
		}
	}
	return errs
}

func (e FieldErrors) ViaIndex(index int) FieldErrors {
	errs := make(FieldErrors, len(e))
	for i, err := range e {
		newField := fmt.Sprintf("[%d]", index)
		if !strings.HasPrefix(err.Field, "[") {
			newField = newField + "."
		}
		if err.Field != "[]" {
			newField = newField + err.Field
		}
		errs[i] = &field.Error{
			Type:     err.Type,
			Field:    newField,
			BadValue: err.BadValue,
			Detail:   err.Detail,
		}
	}
	return errs
}

func (e FieldErrors) ViaFieldIndex(key string, index int) FieldErrors {
	return e.ViaIndex(index).ViaField(key)
}

func (e FieldErrors) ErrorList() field.ErrorList {
	list := make(field.ErrorList, len(e))
	for i := range e {
		list[i] = e[i]
	}
	return list
}

func (e FieldErrors) ToAggregate() error {
	return e.ErrorList().ToAggregate()
}

type Validatable = interface {
	Validate(context.Context) FieldErrors
}

func ErrDisallowedFields(name string, detail string) FieldErrors {
	return FieldErrors{
		field.Forbidden(field.NewPath(name), detail),
	}
}

func ErrInvalidArrayValue(value interface{}, name string, index int) FieldErrors {
	return FieldErrors{
		field.Invalid(field.NewPath(name).Index(index), value, ""),
	}
}

func ErrInvalidValue(value interface{}, name string) FieldErrors {
	return FieldErrors{
		field.Invalid(field.NewPath(name), value, ""),
	}
}

func ErrDuplicateValue(value interface{}, names ...string) FieldErrors {
	errs := FieldErrors{}

	for _, name := range names {
		errs = append(errs, field.Duplicate(field.NewPath(name), value))
	}

	return errs
}

func ErrMissingField(name string) FieldErrors {
	return FieldErrors{
		field.Required(field.NewPath(name), ""),
	}
}

func ErrMissingOneOf(names ...string) FieldErrors {
	return FieldErrors{
		field.Required(field.NewPath(fmt.Sprintf("[%s]", strings.Join(names, ", "))), "expected exactly one, got neither"),
	}
}

func ErrMultipleOneOf(names ...string) FieldErrors {
	return FieldErrors{
		field.Required(field.NewPath(fmt.Sprintf("[%s]", strings.Join(names, ", "))), "expected exactly one, got both"),
	}
}
