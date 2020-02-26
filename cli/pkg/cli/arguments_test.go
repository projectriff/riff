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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/cli/pkg/cli"
	"github.com/spf13/cobra"
)

func TestArgs(t *testing.T) {
	tests := []struct {
		name  string
		items []cli.Arg
		args  []string
		err   error
		fmt   string
	}{{
		name: "no args",
		fmt:  "",
	}, {
		name: "single arity",
		args: []string{"my-arg"},
		items: []cli.Arg{
			{
				Name:  "arg1",
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
		fmt: " <arg1>",
	}, {
		name: "missing args",
		args: []string{},
		items: []cli.Arg{
			{
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					return fmt.Errorf("should not be called")
				},
			},
		},
		err: fmt.Errorf("missing required argument(s)"),
		fmt: "",
	}, {
		name: "extra args",
		args: []string{"my-arg-1", "my-arg-2"},
		items: []cli.Arg{
			{
				Name:  "arg1",
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg-1", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
		err: fmt.Errorf("unknown command %q for %q", "my-arg-2", "args-test"),
		fmt: " <arg1>",
	}, {
		name: "multiple single arity",
		args: []string{"my-arg", "other-arg"},
		items: []cli.Arg{
			{
				Name:  "arg1",
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
			{
				Name:  "arg2",
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("other-arg", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
		fmt: " <arg1> <arg2>",
	}, {
		name: "capture arity",
		args: []string{"my-arg-1", "my-arg-2"},
		items: []cli.Arg{
			{
				Name:  "arg1",
				Arity: -1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff([]string{"my-arg-1", "my-arg-2"}, args[offset:]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
		fmt: " <arg1>",
	}, {
		name: "capture arity, no args",
		args: []string{},
		items: []cli.Arg{
			{
				Name:  "arg1",
				Arity: -1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff([]string{}, args[offset:]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
		fmt: " <arg1>",
	}, {
		name: "capture arity, after single arity",
		args: []string{"my-arg-1", "my-arg-2"},
		items: []cli.Arg{
			{
				Name:  "arg1",
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg-1", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
			{
				Name:  "arg2",
				Arity: -1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff([]string{"my-arg-2"}, args[offset:]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
		fmt: " <arg1> <arg2>",
	}, {
		name: "optional args",
		args: []string{},
		items: []cli.Arg{
			{
				Name:     "arg1",
				Arity:    1,
				Optional: true,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					return fmt.Errorf("should not be called")
				},
			},
		},
		fmt: " [arg1]",
	}, {
		name: "ignored args",
		args: []string{"my-arg"},
		items: []cli.Arg{
			{
				Name:  "arg1",
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					return cli.ErrIgnoreArg
				},
			},
			{
				Name:  "arg2",
				Arity: 1,
				Set: func(cmd *cobra.Command, args []string, offset int) error {
					if diff := cmp.Diff("my-arg", args[offset]); diff != "" {
						return fmt.Errorf("unexpected arg (-expected, +actual): %s", diff)
					}
					return nil
				},
			},
		},
		fmt: " <arg1> <arg2>",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "args-test",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cli.Args(cmd,
				test.items...,
			)
			cmd.SetArgs(test.args)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if expected, actual := test.fmt, cli.FormatArgs(cmd); expected != actual {
				t.Errorf("Expected format %q, actually %q", expected, actual)
			}
		})
	}
}

func TestNameArg(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		actual   string
		expected string
		err      error
	}{{
		name: "too few args",
		err:  fmt.Errorf("missing required argument(s)"),
	}, {
		name:     "name arg",
		args:     []string{"my-name"},
		expected: "my-name",
	}, {
		name:     "too many args",
		args:     []string{"my-name", "extra-arg"},
		expected: "my-name",
		err:      fmt.Errorf("unknown command %q for %q", "extra-arg", "args-test"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "args-test",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cli.Args(cmd,
				cli.NameArg(&test.actual),
			)
			cmd.SetArgs(test.args)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if diff := cmp.Diff(test.expected, test.actual); diff != "" {
				t.Errorf("Unexpected arg binding (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestNamesArg(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		actual   []string
		expected []string
		err      error
	}{{
		name:     "no name",
		args:     []string{},
		expected: []string{},
	}, {
		name:     "single name",
		args:     []string{"my-name"},
		expected: []string{"my-name"},
	}, {
		name:     "multiple names",
		args:     []string{"my-name", "my-other-name"},
		expected: []string{"my-name", "my-other-name"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "args-test",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cli.Args(cmd,
				cli.NamesArg(&test.actual),
			)
			cmd.SetArgs(test.args)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if diff := cmp.Diff(test.expected, test.actual); diff != "" {
				t.Errorf("Unexpected arg binding (-expected, +actual): %s", diff)
			}
		})
	}
}

func TestBareDoubleDashArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		actual   []string
		expected []string
		err      error
	}{{
		name: "no args",
		args: []string{},
	}, {
		name: "no bare double dash",
		args: []string{"my-arg", "my-other-arg"},
	}, {
		name:     "no args after bare double dash",
		args:     []string{"my-arg", "my-other-arg", "--"},
		expected: []string{},
	}, {
		name:     "bare double dash",
		args:     []string{"my-arg", "my-other-arg", "--", "my-name", "my-other-name"},
		expected: []string{"my-name", "my-other-name"},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "args-test",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			cli.Args(cmd,
				cli.BareDoubleDashArgs(&test.actual),
			)
			cmd.SetArgs(test.args)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := cmd.Execute()

			if expected, actual := fmt.Sprintf("%s", test.err), fmt.Sprintf("%s", err); expected != actual {
				t.Errorf("Expected error %q, actually %q", expected, actual)
			}
			if diff := cmp.Diff(test.expected, test.actual); diff != "" {
				t.Errorf("Unexpected arg binding (-expected, +actual): %s", diff)
			}
		})
	}
}
