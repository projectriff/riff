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

	"github.com/google/go-cmp/cmp"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/projectriff/riff/pkg/cli"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func TestPrintResourceStatusForReady(t *testing.T) {
	stdout := &bytes.Buffer{}
	config := &cli.Config{
		Stdout: stdout,
	}
	condition := duckv1alpha1.Condition{
		Type:    buildv1alpha1.FunctionConditionReady,
		Status:  v1.ConditionTrue,
	}
	expected := strings.TrimSpace(`
# test: Ready
---
lastTransitionTime: null
status: "True"
type: Ready`)
	cli.PrintResourceStatus(config, "test", &condition)
	actual := strings.TrimSpace(stdout.String())
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Unexpected stdout (-expected, +actual): %s", diff)
	}
}

func TestPrintResourceStatusForFailure(t *testing.T) {
	stdout := &bytes.Buffer{}
	config := &cli.Config{
		Stdout: stdout,
	}
	condition := duckv1alpha1.Condition{
		Type:    buildv1alpha1.FunctionConditionReady,
		Status:  v1.ConditionFalse,
		Reason: "Failure",
		Severity: "Severe",
		Message: "message that things aren't working",
	}
	expected := strings.TrimSpace(`
# test: Failure
---
lastTransitionTime: null
message: message that things aren't working
reason: Failure
severity: Severe
status: "False"
type: Ready`)
	cli.PrintResourceStatus(config, "test", &condition)
	actual := strings.TrimSpace(stdout.String())
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Unexpected stdout (-expected, +actual): %s", diff)
	}
}
