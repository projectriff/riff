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

package cli_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/cli/pkg/cli"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestFieldErrors_Also(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{Field: "field1"},
		&field.Error{Field: "field2"},
		&field.Error{Field: "field3"},
	}
	actual := cli.FieldErrors{}.Also(
		cli.FieldErrors{
			&field.Error{Field: "field1"},
			&field.Error{Field: "field2"},
		},
		cli.FieldErrors{
			&field.Error{Field: "field3"},
		},
	)

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestFieldErrors_ViaField(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{Field: "parent"},
		&field.Error{Field: "parent.field"},
		&field.Error{Field: "parent[0]"},
	}
	actual := cli.FieldErrors{
		&field.Error{Field: "[]"},
		&field.Error{Field: "field"},
		&field.Error{Field: "[0]"},
	}.ViaField("parent")

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestFieldErrors_ViaIndex(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{Field: "[2]"},
		&field.Error{Field: "[2].field"},
		&field.Error{Field: "[2][0]"},
	}
	actual := cli.FieldErrors{
		&field.Error{Field: "[]"},
		&field.Error{Field: "field"},
		&field.Error{Field: "[0]"},
	}.ViaIndex(2)

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestFieldErrors_ViaFieldIndex(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{Field: "parent[2]"},
		&field.Error{Field: "parent[2].field"},
		&field.Error{Field: "parent[2][0]"},
	}
	actual := cli.FieldErrors{
		&field.Error{Field: "[]"},
		&field.Error{Field: "field"},
		&field.Error{Field: "[0]"},
	}.ViaFieldIndex("parent", 2)

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestFieldErrors_ErrorList(t *testing.T) {
	expected := field.ErrorList{
		&field.Error{Field: "[]"},
	}
	actual := cli.FieldErrors{
		&field.Error{Field: "[]"},
	}.ErrorList()

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestFieldErrors_ToAggregate(t *testing.T) {
	expected := field.ErrorList{
		&field.Error{Field: "[]"},
	}.ToAggregate()
	actual := cli.FieldErrors{
		&field.Error{Field: "[]"},
	}.ToAggregate()

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestErrDisallowedFields(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{
			Type:     field.ErrorTypeForbidden,
			Field:    rifftesting.TestField,
			BadValue: "",
			Detail:   "",
		},
	}
	actual := cli.ErrDisallowedFields(rifftesting.TestField, "")

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestErrInvalidArrayValue(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{
			Type:     field.ErrorTypeInvalid,
			Field:    "test-field[1]",
			BadValue: "value",
			Detail:   "",
		},
	}
	actual := cli.ErrInvalidArrayValue("value", rifftesting.TestField, 1)

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestErrInvalidValue(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{
			Type:     field.ErrorTypeInvalid,
			Field:    "test-field",
			BadValue: "value",
			Detail:   "",
		},
	}
	actual := cli.ErrInvalidValue("value", rifftesting.TestField)

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestErrMissingField(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{
			Type:     field.ErrorTypeRequired,
			Field:    "test-field",
			BadValue: "",
			Detail:   "",
		},
	}
	actual := cli.ErrMissingField(rifftesting.TestField)

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestErrMissingOneOf(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{
			Type:     field.ErrorTypeRequired,
			Field:    "[field1, field2, field3]",
			BadValue: "",
			Detail:   "expected exactly one, got neither",
		},
	}
	actual := cli.ErrMissingOneOf("field1", "field2", "field3")

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestErrMultipleOneOf(t *testing.T) {
	expected := cli.FieldErrors{
		&field.Error{
			Type:     field.ErrorTypeRequired,
			Field:    "[field1, field2, field3]",
			BadValue: "",
			Detail:   "expected exactly one, got both",
		},
	}
	actual := cli.ErrMultipleOneOf("field1", "field2", "field3")

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}
