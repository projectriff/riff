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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/cli/pkg/cli"
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

	// ExpectFieldErrors are the errors that should be returned from the validator.
	ExpectFieldErrors cli.FieldErrors

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

		errs := otr.Options.Validate(context.Background())

		if otr.ExpectFieldErrors != nil {
			actual := errs
			expected := otr.ExpectFieldErrors
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("Unexpected errors (-expected, +actual): %s", diff)
			}
		}

		if expected, actual := otr.ShouldValidate, len(errs) == 0; expected != actual {
			if expected {
				t.Errorf("expected options to validate, actual %q", errs)
			} else {
				t.Errorf("expected options not to validate, actual %q", errs)
			}
		}

		if otr.ShouldValidate == false && len(otr.ExpectFieldErrors) == 0 {
			t.Error("one of ShouldValidate=true or ExpectFieldErrors is required")
		}
	})
}
