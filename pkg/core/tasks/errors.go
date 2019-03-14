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

package tasks

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
