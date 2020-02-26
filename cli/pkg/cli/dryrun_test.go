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

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
)

func TestDryRunResource(t *testing.T) {
	stdout := &bytes.Buffer{}
	ctx := withStdout(context.Background(), stdout)
	resource := &buildv1alpha1.Application{}

	DryRunResource(ctx, resource, resource.GetGroupVersionKind())

	expected := strings.TrimSpace(`
---
apiVersion: build.projectriff.io/v1alpha1
kind: Application
metadata:
  creationTimestamp: null
spec:
  build:
    resources: {}
  image: ""
status: {}
`)
	actual := strings.TrimSpace(stdout.String())
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Unexpected stdout (-expected, +actual): %s", diff)
	}

}
