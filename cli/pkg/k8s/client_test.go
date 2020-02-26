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

package k8s_test

import (
	"testing"

	"github.com/projectriff/cli/pkg/k8s"
)

func TestNewClient(t *testing.T) {
	client := k8s.NewClient("testdata/.kube/config")

	if expected, actual := "my-namespace", client.DefaultNamespace(); expected != actual {
		t.Errorf("Expected namespace to be %q, actually %q", expected, actual)
	}
	if client.KubeRestConfig() == nil {
		t.Errorf("Expected REST config to not be nil")
	}
	if client.Core() == nil {
		t.Errorf("Expected Core client to not be nil")
	}
	if client.Auth() == nil {
		t.Errorf("Expected Auth client to not be nil")
	}
	if client.APIExtension() == nil {
		t.Errorf("Expected APIExtension client to not be nil")
	}
	if client.Build() == nil {
		t.Errorf("Expected Build client to not be nil")
	}
	if client.CoreRuntime() == nil {
		t.Errorf("Expected CoreRuntime client to not be nil")
	}
	if client.StreamingRuntime() == nil {
		t.Errorf("Expected StreamingRuntime client to not be nil")
	}
	if client.KnativeRuntime() == nil {
		t.Errorf("Expected KnativeRuntime client to not be nil")
	}
}
