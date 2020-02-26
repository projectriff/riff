package validation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/validation"
)

func TestValidMimeType(t *testing.T) {
	expected := cli.FieldErrors{}
	actual := validation.MimeType("text/csv", "some-field")

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestInvalidMimeTypeWithMissingSlash(t *testing.T) {
	expected := cli.ErrInvalidValue("invalid", "some-field")
	actual := validation.MimeType("invalid", "some-field")

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}

func TestInvalidMimeTypeWithSingleTrailingSlash(t *testing.T) {
	expected := cli.ErrInvalidValue("invalid/", "some-field")
	actual := validation.MimeType("invalid/", "some-field")

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected, +actual): %s", diff)
	}
}
