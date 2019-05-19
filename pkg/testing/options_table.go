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
	"reflect"
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/pkg/cli"
)

type OptionsTable []OptionsTableRecord

type OptionsTableRecord struct {
	// Name is used to identify the record in the test results. A sub-test is created for each
	// record with this name.
	Name string
	// Skip suppresses the execution of this test record.
	Skip bool
	// Focus executes this record skipping all unfocused records. The containing test will fail to
	// prevent accidental check-in.
	Focus bool

	// inputs

	// Options to validate
	Options cli.Validatable

	// outputs

	// ExpectFieldError is the error that should be returned from the validation.
	ExpectFieldError *cli.FieldError

	// ShouldValidate is true if the options are valid
	ShouldValidate bool
}

func (ot OptionsTable) Run(t *testing.T) {
	focusedTable := OptionsTable{}
	for _, otr := range ot {
		if otr.Focus == true && otr.Skip != true {
			focusedTable = append(focusedTable, otr)
		}
	}
	if len(focusedTable) != 0 {
		for _, otr := range focusedTable {
			otr.Run(t)
		}
		t.Errorf("test run focused on %d record(s), skipped %d record(s)", len(focusedTable), len(ot)-len(focusedTable))
		return
	}

	for _, otr := range ot {
		otr.Run(t)
	}
}

func (otr OptionsTableRecord) Run(t *testing.T) {
	t.Run(otr.Name, func(t *testing.T) {
		if otr.Skip {
			t.SkipNow()
		}

		errs := otr.Options.Validate(context.TODO())
		if errs == nil {
			errs = &cli.FieldError{}
		}

		if otr.ExpectFieldError != nil {
			actual := flattenFieldErrors(errs)
			expected := flattenFieldErrors(otr.ExpectFieldError)
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

		if otr.ShouldValidate == false && otr.ExpectFieldError == nil {
			t.Error("one of ShouldValidate=true or ExpectFieldError is required")
		}
	})
}

var compareFieldError = cmp.Comparer(func(a, b cli.FieldError) bool {
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

func flattenFieldErrors(err *cli.FieldError) []cli.FieldError {
	errs := []cli.FieldError{}

	if err.Message != "" {
		errs = append(errs, *err)
	}
	for _, nestedErr := range extractNestedErrors(err) {
		errs = append(errs, flattenFieldErrors(&nestedErr)...)
	}

	return errs
}

func extractNestedErrors(err *cli.FieldError) []cli.FieldError {
	var nestedErrors []cli.FieldError

	// `nestedErrors = err.errors`
	// TODO let's get this exposed on the type so we don't need to do unsafe reflection
	ev := reflect.ValueOf(err).Elem().FieldByName("errors")
	ev = reflect.NewAt(ev.Type(), unsafe.Pointer(ev.UnsafeAddr())).Elem()
	nev := reflect.ValueOf(&nestedErrors).Elem()
	nev = reflect.NewAt(nev.Type(), unsafe.Pointer(nev.UnsafeAddr())).Elem()
	nev.Set(ev)

	return nestedErrors
}
