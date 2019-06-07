package validation_test

import (
	"github.com/projectriff/riff/pkg/validation"
	"testing"
)

func TestValidMimeType(t *testing.T) {
	err := validation.MimeType("text/csv", "some-field")

	actualErrorMessage := err.Error()
	if actualErrorMessage != "" {
		t.Fatalf("Expected no error, got %q", actualErrorMessage)
	}
}

func TestInvalidMimeTypeWithMissingSlash(t *testing.T) {
	expectedErrorMessage := "invalid value: invalid: some-field"
	err := validation.MimeType("invalid", "some-field")

	actualErrorMessage := err.Error()
	if actualErrorMessage != expectedErrorMessage {
		t.Fatalf("Expected %q as error message, got %q", expectedErrorMessage, actualErrorMessage)
	}
}

func TestInvalidMimeTypeWithSingleTrailingSlash(t *testing.T) {
	expectedErrorMessage := "invalid value: invalid/: some-field"
	err := validation.MimeType("invalid/", "some-field")

	actualErrorMessage := err.Error()
	if actualErrorMessage != expectedErrorMessage {
		t.Fatalf("Expected %q as error message, got %q", expectedErrorMessage, actualErrorMessage)
	}
}