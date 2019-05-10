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

package testing

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/knative/pkg/apis"
)

type OptionsTable []OptionsTableRecord

type OptionsTableRecord struct {
	Name       string
	Skip       bool
	Focus      bool
	Sequential bool

	// inputs
	OverrideOptions interface{}

	// outputs
	ExpectErrorPaths []string
	ExpectErrors     []apis.FieldError
	ShouldValidate   bool
	Verify           func(t *T, err *apis.FieldError)
}

func (ot OptionsTable) Run(t *T, defaultOptionsFactory func() apis.Validatable) {
	focusedTable := OptionsTable{}
	for _, otr := range ot {
		if otr.Focus == true && otr.Skip != true {
			focusedTable = append(focusedTable, otr)
		}
	}
	if len(focusedTable) != 0 {
		for _, otr := range focusedTable {
			otr.Run(t, defaultOptionsFactory)
		}
		t.Errorf("test run focused on %d record(s), skipped %d record(s)", len(focusedTable), len(ot)-len(focusedTable))
		return
	}

	for _, otr := range ot {
		otr.Run(t, defaultOptionsFactory)
	}
}

func (otr OptionsTableRecord) Run(t *T, defaultOptionsFactory func() apis.Validatable) {
	t.Run(otr.Name, func(t *T) {
		if otr.Skip {
			t.SkipNow()
		}
		if !otr.Sequential {
			t.Parallel()
		}

		opts := defaultOptionsFactory()
		if otr.OverrideOptions != nil {
			oov := reflect.ValueOf(otr.OverrideOptions)
			if !isOverideOptionsFunc(oov.Type()) {
				panic(fmt.Sprintf("invalid override options function: %T", otr.OverrideOptions))
			}
			oov.Call([]reflect.Value{reflect.ValueOf(opts)})
		}

		errs := opts.Validate(context.TODO())
		if errs == nil {
			errs = &apis.FieldError{}
		}

		if otr.ExpectErrorPaths != nil {
			actual := flattenFieldPaths(errs)
			expected := otr.ExpectErrorPaths
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("Unexpected field paths (-expected, +actual): %s", diff)
			}
		}

		if otr.ExpectErrors != nil {
			actual := flattenFieldErrors(errs)
			expected := otr.ExpectErrors
			if diff := cmp.Diff(expected, actual, compareFieldError); diff != "" {
				t.Errorf("Unexpected errors (-expected, +actual): %s", diff)
			}
		}

		if expected, actual := otr.ShouldValidate, errs.Error() == ""; expected != actual {
			if expected {
				t.Errorf("expected options to validate, actual %q", errs)
			} else {
				t.Errorf("expected options not to validate, actual %q", errs)
			}
		}

		if otr.ShouldValidate == false && otr.ExpectErrorPaths == nil && otr.ExpectErrors == nil {
			t.Error("at least one of ShouldValidate=true, ExpectErrorPaths or ExpectErrors is required")
		}

		if otr.Verify != nil {
			otr.Verify(t, errs)
		}
	})
}

var compareFieldError = cmp.Comparer(func(a, b apis.FieldError) bool {
	if a.Message != b.Message {
		return false
	}
	if a.Details != b.Details {
		return false
	}
	return cmp.Equal(filterEmpty(a.Paths), filterEmpty(b.Paths))
})

func filterEmpty(s []string) []string {
	r := []string{}
	for _, i := range s {
		if i != "" {
			r = append(r, i)
		}
	}
	return r
}

type OverrideOptionsFunc func(apis.Validatable)

func isOverideOptionsFunc(t reflect.Type) bool {
	if t == nil || t.Kind() != reflect.Func || t.IsVariadic() {
		return false
	}
	if t.NumIn() == 1 && t.NumOut() == 0 && t.In(0).ConvertibleTo(reflect.TypeOf((*apis.Validatable)(nil)).Elem()) {
		return true
	}
	return false
}

func flattenFieldPaths(err *apis.FieldError) []string {
	paths := err.Paths
	if paths == nil {
		paths = []string{}
	}

	for _, nestedErr := range extractNestedErrors(err) {
		paths = append(paths, flattenFieldPaths(&nestedErr)...)
	}

	return paths
}

func flattenFieldErrors(err *apis.FieldError) []apis.FieldError {
	errs := []apis.FieldError{}

	if err.Message != "" {
		errs = append(errs, *err)
	}
	for _, nestedErr := range extractNestedErrors(err) {
		errs = append(errs, flattenFieldErrors(&nestedErr)...)
	}

	return errs
}

func extractNestedErrors(err *apis.FieldError) []apis.FieldError {
	var nestedErrors []apis.FieldError

	// `nestedErrors = err.errors`
	// TODO let's get this exposed on the type so we don't need to do unsafe reflection
	ev := reflect.ValueOf(err).Elem().FieldByName("errors")
	ev = reflect.NewAt(ev.Type(), unsafe.Pointer(ev.UnsafeAddr())).Elem()
	nev := reflect.ValueOf(&nestedErrors).Elem()
	nev = reflect.NewAt(nev.Type(), unsafe.Pointer(nev.UnsafeAddr())).Elem()
	nev.Set(ev)

	return nestedErrors
}
