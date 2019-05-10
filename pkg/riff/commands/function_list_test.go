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
	"fmt"
	"strings"

	"github.com/knative/pkg/apis"
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFunctionListOptions(t *testing.T) {
	defaultOptions := func() testing.Validatable {
		return &commands.FunctionListOptions{
			Namespace: "default",
		}
	}

	table := testing.OptionsTable{
		{
			Name:           "default",
			ShouldValidate: true,
		},
		{
			Name: "custom namespace",
			OverrideOptions: func(opts *commands.FunctionListOptions) {
				opts.Namespace = "my-namespace"
			},
			ShouldValidate: true,
		},
		{
			Name: "all namespaces",
			OverrideOptions: func(opts *commands.FunctionListOptions) {
				opts.Namespace = ""
				opts.AllNamespaces = true
			},
			ShouldValidate: true,
		},
		{
			Name: "neither",
			OverrideOptions: func(opts *commands.FunctionListOptions) {
				opts.Namespace = ""
			},
			ExpectErrors: []apis.FieldError{
				*apis.ErrMissingOneOf("namespace", "all-namespaces"),
			},
		},
		{
			Name: "both",
			OverrideOptions: func(opts *commands.FunctionListOptions) {
				opts.AllNamespaces = true
			},
			ExpectErrors: []apis.FieldError{
				*apis.ErrMultipleOneOf("namespace", "all-namespaces"),
			},
		},
	}

	table.Run(t, defaultOptions)
}

func TestFunctionListCommand(t *testing.T) {
	t.Parallel()

	functionName := "test-function"
	functionAltName := "test-alt-function"
	defaultNamespace := "default"
	altNamespace := "alt-namespace"

	table := testing.CommandTable{
		{
			Name: "empty",
			Args: []string{},
			Verify: func(t *testing.T, output string, err error) {
				if expected, actual := output, "No functions found.\n"; actual != expected {
					t.Errorf("expected output %q, actually %q", expected, actual)
				}
			},
		},
		{
			Name: "lists a secret",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionName,
						Namespace: defaultNamespace,
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if actual, want := output, fmt.Sprintf("%s\n", functionName); actual != want {
					t.Errorf("expected output %q, actually %q", want, actual)
				}
			},
		},
		{
			Name: "filters by namespace",
			Args: []string{"--namespace", altNamespace},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionName,
						Namespace: defaultNamespace,
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if actual, want := output, "No functions found.\n"; actual != want {
					t.Errorf("expected output %q, actually %q", want, actual)
				}
			},
		},
		{
			Name: "all namespace",
			Args: []string{"--all-namespaces"},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionName,
						Namespace: defaultNamespace,
					},
				},
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionAltName,
						Namespace: altNamespace,
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					fmt.Sprintf("%s\n", functionName),
					fmt.Sprintf("%s\n", functionAltName),
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("list", "functions"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewFunctionListCommand)
}
