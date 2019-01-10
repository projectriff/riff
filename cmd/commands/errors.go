package commands

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

// Joins and formats results.
// Returns nil if all errors are nil.
// If there is only one non-nil error, the  error is returned as-is.
func MergeResults(input []error) error {
	stats := computeStats(input)
	length := stats.nilFreeSliceLength
	if length == 0 {
		return nil
	}
	if length == 1 {
		return input[stats.firstNonNilIndex]
	}
	return errors.New(formatErrorMessages(input))
}

func formatErrorMessages(input []error) string {
	mergedMessage := strings.Builder{}
	for i, err := range input {
		if err == nil {
			mergedMessage.WriteString(fmt.Sprintf("# Call %d - no error\n", i+1))
			continue
		}
		mergedMessage.WriteString(fmt.Sprintf("# Call %d - error\n", i+1))
		leftPaddedMessage := strings.Replace(err.Error(), "\n", "\n\t", -1)
		mergedMessage.WriteString("\t" + leftPaddedMessage + "\n")
	}
	s := mergedMessage.String()
	return s
}

func computeStats(errors []error) stats {
	firstNonNilIndex := -1
	length := 0
	for i, err := range errors {
		if err != nil {
			length++
			if firstNonNilIndex == -1 {
				firstNonNilIndex = i
			}
		}
	}
	return stats{
		nilFreeSliceLength: length,
		firstNonNilIndex:   firstNonNilIndex,
	}
}

type stats struct {
	firstNonNilIndex   int
	nilFreeSliceLength int
}
