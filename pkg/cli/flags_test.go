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
	"fmt"
	"testing"

	"github.com/projectriff/riff/pkg/cli"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	"github.com/spf13/cobra"
)

func TestAllNamespacesFlag(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		prior               func(cmd *cobra.Command, args []string) error
		namespace           string
		actualNamespace     string
		allNamespaces       bool
		actualAllNamespaces bool
		err                 error
	}{{
		name:      "default",
		args:      []string{},
		namespace: "default",
	}, {
		name:      "explicit namespace",
		args:      []string{cli.NamespaceFlagName, "my-namespace"},
		namespace: "my-namespace",
	}, {
		name:          "all namespaces",
		args:          []string{cli.AllNamespacesFlagName},
		allNamespaces: true,
	}, {
		name: "explicit namespace and all namespaces",
		args: []string{cli.NamespaceFlagName, "default", cli.AllNamespacesFlagName},
		err:  cli.ErrMultipleOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName),
	}, {
		name: "prior PreRunE",
		args: []string{},
		prior: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		namespace: "default",
	}, {
		name: "prior PreRunE, error",
		args: []string{},
		prior: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("prior PreRunE error")
		},
		err: fmt.Errorf("prior PreRunE error"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := cli.NewDefaultConfig()
			c.Client = rifftesting.NewClient()
			cmd := &cobra.Command{
				PreRunE: test.prior,
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}

			cli.AllNamespacesFlag(cmd, c, &test.actualNamespace, &test.actualAllNamespaces)

			if cmd.Flag(cli.StripDash(cli.NamespaceFlagName)) == nil {
				t.Errorf("Expected %s to be defined", cli.NamespaceFlagName)
			}
			if cmd.Flag(cli.StripDash(cli.AllNamespacesFlagName)) == nil {
				t.Errorf("Expected %s to be defined", cli.AllNamespacesFlagName)
			}

			cmd.SetArgs(test.args)
			cmd.SetOutput(&bytes.Buffer{})
			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if err == nil {
				if expected, actual := test.namespace, test.actualNamespace; expected != actual {
					t.Errorf("Expected namespace %q, actually %q", expected, actual)
				}
				if expected, actual := test.allNamespaces, test.actualAllNamespaces; expected != actual {
					t.Errorf("Expected all namespace %v, actually %v", expected, actual)
				}
			}
		})
	}
}

func TestNamespaceFlag(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		prior           func(cmd *cobra.Command, args []string) error
		namespace       string
		actualNamespace string
		err             error
	}{{
		name:      "default",
		args:      []string{},
		namespace: "default",
	}, {
		name:      "explicit namespace",
		args:      []string{cli.NamespaceFlagName, "my-namespace"},
		namespace: "my-namespace",
	}, {
		name: "prior PreRunE",
		args: []string{},
		prior: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		namespace: "default",
	}, {
		name: "prior PreRunE, error",
		args: []string{},
		prior: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("prior PreRunE error")
		},
		err: fmt.Errorf("prior PreRunE error"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := cli.NewDefaultConfig()
			c.Client = rifftesting.NewClient()
			cmd := &cobra.Command{
				PreRunE: test.prior,
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}

			cli.NamespaceFlag(cmd, c, &test.actualNamespace)

			if cmd.Flag(cli.StripDash(cli.NamespaceFlagName)) == nil {
				t.Errorf("Expected %s to be defined", cli.NamespaceFlagName)
			}
			if cmd.Flag(cli.StripDash(cli.AllNamespacesFlagName)) != nil {
				t.Errorf("Expected %s not to be defined", cli.AllNamespacesFlagName)
			}

			cmd.SetArgs(test.args)
			cmd.SetOutput(&bytes.Buffer{})
			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if err == nil {
				if expected, actual := test.namespace, test.actualNamespace; expected != actual {
					t.Errorf("Expected namespace %q, actually %q", expected, actual)
				}
			}
		})
	}
}

func TestStripDash(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{{
		name: "empty",
	}, {
		name:   "remove leading dash",
		input:  "--flag",
		output: "flag",
	}, {
		name:   "ingore extra dash",
		input:  "--flag-name",
		output: "flag-name",
	}, {
		name:   "ingore extra doubledash",
		input:  "--flag--",
		output: "flag--",
	}, {
		name:   "no dashes",
		input:  "flag",
		output: "flag",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if expected, actual := test.output, cli.StripDash(test.input); expected != actual {
				t.Errorf("Expected dash stripped string to be %q, actually %q", expected, actual)
			}
		})
	}
}
