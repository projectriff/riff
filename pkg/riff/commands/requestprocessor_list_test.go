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

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRequestProcessorListOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.RequestProcessorListOptions{
				ListOptions: testing.InvalidListOptions,
			},
			ExpectFieldError: testing.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.RequestProcessorListOptions{
				ListOptions: testing.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestRequestProcessorListCommand(t *testing.T) {
	requestprocessorsName := "test-requestprocessors"
	requestprocessorOtherName := "test-other-requestprocessors"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"

	table := testing.CommandTable{
		{
			Name: "invalid args",
			Args: []string{},
			Prepare: func(t *testing.T, c *cli.Config) error {
				// disable default namespace
				c.Client.(*testing.FakeClient).Namespace = ""
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "empty",
			Args: []string{},
			Verify: func(t *testing.T, output string, err error) {
				if expected, actual := output, "No request processors found.\n"; actual != expected {
					t.Errorf("expected output %q, actually %q", expected, actual)
				}
			},
		},
		{
			Name: "lists a secret",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorsName,
						Namespace: defaultNamespace,
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if actual, want := output, fmt.Sprintf("%s\n", requestprocessorsName); actual != want {
					t.Errorf("expected output %q, actually %q", want, actual)
				}
			},
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorsName,
						Namespace: defaultNamespace,
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if actual, want := output, "No request processors found.\n"; actual != want {
					t.Errorf("expected output %q, actually %q", want, actual)
				}
			},
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorsName,
						Namespace: defaultNamespace,
					},
				},
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					fmt.Sprintf("%s\n", requestprocessorsName),
					fmt.Sprintf("%s\n", requestprocessorOtherName),
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
				testing.InduceFailure("list", "requestprocessors"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewRequestProcessorListCommand)
}
