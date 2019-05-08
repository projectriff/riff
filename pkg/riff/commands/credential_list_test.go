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

package commands_test

import (
	"strings"

	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCredentialListCommand(t *testing.T) {
	testing.Table{{
		Name: "empty",
		Args: []string{},
		WithOutput: func(t *testing.T, output string, err error) {
			if got, want := output, "No credentials found.\n"; got != want {
				t.Errorf("expected output %q got %q", want, got)
			}
		},
	}, {
		Name: "lists a secret",
		Args: []string{},
		Objects: []runtime.Object{
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-secret",
					Namespace: "default",
				},
			},
		},
		WithOutput: func(t *testing.T, output string, err error) {
			if got, want := output, "my-secret\n"; got != want {
				t.Errorf("expected output %q got %q", want, got)
			}
		},
	}, {
		Name: "filters by namespace",
		Args: []string{"--namespace", "my-namespace"},
		Objects: []runtime.Object{
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-secret",
					Namespace: "default",
				},
			},
		},
		WithOutput: func(t *testing.T, output string, err error) {
			if got, want := output, "No credentials found.\n"; got != want {
				t.Errorf("expected output %q got %q", want, got)
			}
		},
	}, {
		Name: "all namespace",
		Args: []string{"--all-namespaces"},
		Objects: []runtime.Object{
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-secret",
					Namespace: "default",
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-other-secret",
					Namespace: "my-namespace",
				},
			},
		},
		WithOutput: func(t *testing.T, got string, err error) {
			for _, want := range []string{
				"my-secret\n",
				"my-other-secret\n",
			} {
				if !strings.Contains(got, want) {
					t.Errorf("expected command output to contain %q got %q", want, got)
				}
			}
		},
	}, {
		Name: "list error",
		Args: []string{},
		WithReactors: []testing.ReactionFunc{
			testing.InduceFailure("list", "secrets"),
		},
		WantError: true,
	}}.Run(t, commands.NewCredentialListCommand)
}
