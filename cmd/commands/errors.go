package commands

import (
	"github.com/pkg/errors"
	"strings"
)

// Joins and formats results, with the provided formatter.
// The result is skipped if the formatter returns an empty string.
// Returns nil if all errors are nil.
func MergeResults(results []CorrelatedResult, formatter func(CorrelatedResult) string) error {
	stats := computeStats(results)
	length := stats.nilFreeSliceLength
	if length == 0 {
		return nil
	}
	if length == 1 {
		return errors.New(format(formatter, results[stats.firstNonNilIndex]))
	}
	return errors.New(formatErrorMessages(results, formatter))
}

func formatErrorMessages(input []CorrelatedResult, formatter func(CorrelatedResult) string) string {
	mergedMessage := strings.Builder{}
	for _, err := range input {
		formattedMessage := format(formatter, err)
		if formattedMessage != "" {
			mergedMessage.WriteString(formattedMessage)
			mergedMessage.WriteString("\n")
		}
	}
	return strings.TrimSuffix(mergedMessage.String(), "\n")
}

func format(formatter func(CorrelatedResult) string, err CorrelatedResult) string {
	return formatter(err)
}

func computeStats(results []CorrelatedResult) *stats {
	firstNonNilIndex := -1
	length := 0
	for i, result := range results {
		if result.Error != nil {
			length++
			if firstNonNilIndex == -1 {
				firstNonNilIndex = i
			}
		}
	}
	return &stats{
		nilFreeSliceLength: length,
		firstNonNilIndex:   firstNonNilIndex,
	}
}

type stats struct {
	firstNonNilIndex   int
	nilFreeSliceLength int
}
