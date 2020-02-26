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
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/system/pkg/apis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPrintResourceStatus(t *testing.T) {
	tests := []struct {
		name      string
		condition *apis.Condition
		output    string
	}{{
		name: "nil",
		output: `
# test: <unknown>
`,
	}, {
		name:      "empty",
		condition: &apis.Condition{},
		output: `
# test: <unknown>
---
lastTransitionTime: null
status: ""
type: ""
`,
	}, {
		name: "unknown",
		condition: &apis.Condition{
			Type:    apis.ConditionReady,
			Status:  corev1.ConditionUnknown,
			Reason:  "HangOn",
			Message: "a hopefully informative message about what's in flight",
			LastTransitionTime: apis.VolatileTime{
				Inner: metav1.Time{
					Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
				},
			},
		},
		output: `
# test: Unknown
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what's in flight
reason: HangOn
status: Unknown
type: Ready
`,
	}, {
		name: "ready",
		condition: &apis.Condition{
			Type:   apis.ConditionReady,
			Status: corev1.ConditionTrue,
			LastTransitionTime: apis.VolatileTime{
				Inner: metav1.Time{
					Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
				},
			},
		},
		output: `
# test: Ready
---
lastTransitionTime: "2019-06-29T01:44:05Z"
status: "True"
type: Ready
`,
	}, {
		name: "failure",
		condition: &apis.Condition{
			Type:    apis.ConditionReady,
			Status:  corev1.ConditionFalse,
			Reason:  "OopsieDoodle",
			Message: "a hopefully informative message about what went wrong",
			LastTransitionTime: apis.VolatileTime{
				Inner: metav1.Time{
					Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
				},
			},
		},
		output: `
# test: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			config := &cli.Config{
				Stdout: output,
			}
			cli.PrintResourceStatus(config, "test", test.condition)
			expected, actual := strings.TrimSpace(test.output), strings.TrimSpace(output.String())
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}
