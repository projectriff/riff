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
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
)

func TestSequence(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		items  []func(cmd *cobra.Command, args []string) error
		output string
		err    error
	}{{
		name: "empty",
	}, {
		name: "single item",
		args: []string{"a", "b", "c"},
		items: []func(cmd *cobra.Command, args []string) error{
			func(cmd *cobra.Command, args []string) error {
				fmt.Fprintf(cmd.OutOrStdout(), "step %v\n", args)
				return nil
			},
		},
		output: `
step [a b c]
`,
	}, {
		name: "multiple items",
		items: []func(cmd *cobra.Command, args []string) error{
			func(cmd *cobra.Command, args []string) error {
				fmt.Fprintln(cmd.OutOrStdout(), "step 1")
				return nil
			},
			func(cmd *cobra.Command, args []string) error {
				fmt.Fprintln(cmd.OutOrStdout(), "step 2")
				return nil
			},
			func(cmd *cobra.Command, args []string) error {
				fmt.Fprintln(cmd.OutOrStdout(), "step 3")
				return nil
			},
		},
		output: `
step 1
step 2
step 3
`,
	}, {
		name: "stops on error",
		items: []func(cmd *cobra.Command, args []string) error{
			func(cmd *cobra.Command, args []string) error {
				fmt.Fprintln(cmd.OutOrStdout(), "step 1")
				return nil
			},
			func(cmd *cobra.Command, args []string) error {
				fmt.Fprintln(cmd.OutOrStdout(), "step 2")
				return fmt.Errorf("test error")
			},
			func(cmd *cobra.Command, args []string) error {
				fmt.Fprintln(cmd.OutOrStdout(), "step 3")
				return nil
			},
		},
		output: `
step 1
step 2
`,
		err: fmt.Errorf("test error"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cmd := &cobra.Command{}
			cmd.SetOutput(output)

			err := cli.Sequence(test.items...)(cmd, test.args)

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if diff := cmp.Diff(strings.TrimSpace(test.output), strings.TrimSpace(output.String())); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestReadStdin(t *testing.T) {
	// TODO is it possible to test the IsTerminal branch?
	expected := []byte("hello")
	var actual []byte

	c := cli.NewDefaultConfig()
	c.Stdin = bytes.NewBuffer(expected)
	runE := cli.ReadStdin(c, &actual, "> ")

	cmd := &cobra.Command{}
	args := []string{}

	err := runE(cmd, args)

	if err != nil {
		t.Errorf("expected no error, actually %v", err)
	}
	if string(expected) != string(actual) {
		t.Errorf("expected input %q, actually %q", expected, actual)
	}
}

func TestCommandFromContext_WithCommand(t *testing.T) {
	cmd := &cobra.Command{}
	parentCtx := context.TODO()
	childCtx := cli.WithCommand(parentCtx, cmd)

	if expected, actual := (*cobra.Command)(nil), cli.CommandFromContext(parentCtx); expected != actual {
		t.Errorf("expected command %v, actually %v", expected, actual)
	}
	if expected, actual := cmd, cli.CommandFromContext(childCtx); expected != actual {
		t.Errorf("expected command %v, actually %v", expected, actual)
	}
}
