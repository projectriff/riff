/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
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
	"time"

	"github.com/fatih/color"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/projectriff/riff/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFormatTimestampSince(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	now := time.Now()

	tests := []struct {
		name   string
		input  metav1.Time
		output string
	}{{
		name:   "empty",
		output: cli.Swarnf("<unknown>"),
	}, {
		name:   "now",
		input:  metav1.Time{now},
		output: "0s",
	}, {
		name:   "1 minute ago",
		input:  metav1.Time{now.Add(-1 * time.Minute)},
		output: "60s",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, cli.FormatTimestampSince(test.input, now); expected != actual {
				t.Errorf("Expected formated string to be %q, actually %q", expected, actual)
			}
		})
	}
}

func TestFormatEmptyString(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = noColor }()

	tests := []struct {
		name   string
		input  string
		output string
	}{{
		name:   "empty",
		output: cli.Sfaintf("<empty>"),
	}, {
		name:   "not empty",
		input:  "hello",
		output: "hello",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, cli.FormatEmptyString(test.input); expected != actual {
				t.Errorf("Expected formated string to be %q, actually %q", expected, actual)
			}
		})
	}
}

func TestFormatConditionStatus(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = noColor }()

	tests := []struct {
		name   string
		input  *duckv1alpha1.Condition
		output string
	}{{
		name:   "empty",
		output: cli.Swarnf("<unknown>"),
	}, {
		name: "status true",
		input: &duckv1alpha1.Condition{
			Type:   duckv1alpha1.ConditionReady,
			Status: corev1.ConditionTrue,
		},
		output: cli.Ssuccessf("Ready"),
	}, {
		name: "status false",
		input: &duckv1alpha1.Condition{
			Type:   duckv1alpha1.ConditionReady,
			Status: corev1.ConditionFalse,
			Reason: "uh-oh",
		},
		output: cli.Serrorf("uh-oh"),
	}, {
		name: "status false, no reason",
		input: &duckv1alpha1.Condition{
			Type:   duckv1alpha1.ConditionReady,
			Status: corev1.ConditionFalse,
		},
		output: cli.Serrorf("not-Ready"),
	}, {
		name: "status unknown",
		input: &duckv1alpha1.Condition{
			Type:   duckv1alpha1.ConditionReady,
			Status: corev1.ConditionUnknown,
		},
		output: cli.Sinfof("Unknown"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, cli.FormatConditionStatus(test.input); expected != actual {
				t.Errorf("Expected formated string to be %q, actually %q", expected, actual)
			}
		})
	}
}
