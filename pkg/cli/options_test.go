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

package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/projectriff/riff/pkg/cli"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	"github.com/spf13/cobra"
)

type StubValidateOptions struct {
	called        bool
	validationErr *cli.FieldError
}

func (o *StubValidateOptions) Validate(ctx context.Context) *cli.FieldError {
	o.called = true
	return o.validationErr
}

func TestValidateOptions(t *testing.T) {
	tests := []struct {
		name          string
		opts          *StubValidateOptions
		expectedErr   error
		usageSilenced bool
	}{{
		name:          "valid, no error",
		opts:          &StubValidateOptions{},
		usageSilenced: true,
	}, {
		name: "valid, empty error",
		opts: &StubValidateOptions{
			validationErr: cli.EmptyFieldError,
		},
		usageSilenced: true,
	}, {
		name: "validation error",
		opts: &StubValidateOptions{
			validationErr: cli.ErrMissingField("field-name"),
		},
		expectedErr:   cli.ErrMissingField("field-name"),
		usageSilenced: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			err := cli.ValidateOptions(test.opts)(cmd, []string{})

			if expected, actual := true, test.opts.called; true != actual {
				t.Errorf("expected called to be %v, actually %v", expected, actual)
			}
			if expected, actual := test.expectedErr, err; fmt.Sprintf("%s", expected) != fmt.Sprintf("%s", actual) {
				t.Errorf("expected error to be %v, actually %v", expected, actual)
			}
			if expected, actual := test.usageSilenced, cmd.SilenceUsage; expected != actual {
				t.Errorf("expected cmd.SilenceUsage to be %v, actually %v", expected, actual)
			}
		})
	}
}

type StubExecOptions struct {
	dryRun  bool
	called  bool
	config  *cli.Config
	cmd     *cobra.Command
	execErr error
}

func (o *StubExecOptions) Exec(ctx context.Context, c *cli.Config) error {
	o.called = true
	o.config = c
	o.cmd = cli.CommandFromContext(ctx)
	return o.execErr
}

func (o *StubExecOptions) IsDryRun() bool {
	return o.dryRun
}

func TestExecOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        *StubExecOptions
		expectedErr error
	}{{
		name: "success",
		opts: &StubExecOptions{},
	}, {
		name: "failure",
		opts: &StubExecOptions{
			execErr: fmt.Errorf("test exec error"),
		},
		expectedErr: fmt.Errorf("test exec error"),
	}, {
		name: "dry run",
		opts: &StubExecOptions{
			dryRun: true,
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			config := &cli.Config{
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
			}
			err := cli.ExecOptions(config, test.opts)(cmd, []string{})

			if expected, actual := true, test.opts.called; true != actual {
				t.Errorf("expected called to be %v, actually %v", expected, actual)
			}
			if expected, actual := test.expectedErr, err; fmt.Sprintf("%s", expected) != fmt.Sprintf("%s", actual) {
				t.Errorf("expected error to be %v, actually %v", expected, actual)
			}
			if expected, actual := config, test.opts.config; expected != actual {
				t.Errorf("expected config to be %v, actually %v", expected, actual)
			}
			if expected, actual := cmd, test.opts.cmd; expected != actual {
				t.Errorf("expected command to be %v, actually %v", expected, actual)
			}
			if test.opts.dryRun {
				if config.Stdout != config.Stderr {
					t.Errorf("expected stdout and stderr to be the same, actually %v %v", config.Stdout, config.Stderr)
				}
			} else {
				if config.Stdout == config.Stderr {
					t.Errorf("expected stdout and stderr to be different, actually %v %v", config.Stdout, config.Stderr)
				}
			}
		})
	}
}

func TestListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "default",
			Options: &cli.ListOptions{
				Namespace: "default",
			},
			ShouldValidate: true,
		},
		{
			Name: "all namespaces",
			Options: &cli.ListOptions{
				AllNamespaces: true,
			},
			ShouldValidate: true,
		},
		{
			Name:             "neither",
			Options:          &cli.ListOptions{},
			ExpectFieldError: cli.ErrMissingOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName),
		},
		{
			Name: "both",
			Options: &cli.ListOptions{
				Namespace:     "default",
				AllNamespaces: true,
			},
			ExpectFieldError: cli.ErrMultipleOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName),
		},
	}

	table.Run(t)
}

func TestResourceOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name:    "default",
			Options: &cli.ResourceOptions{},
			ExpectFieldError: cli.EmptyFieldError.Also(
				cli.ErrMissingField(cli.NamespaceFlagName),
				cli.ErrMissingField(cli.NameArgumentName),
			),
		},
		{
			Name: "has both",
			Options: &cli.ResourceOptions{
				Namespace: "default",
				Name:      "push-credentials",
			},
			ShouldValidate: true,
		},
		{
			Name: "missing namespace",
			Options: &cli.ResourceOptions{
				Name: "push-credentials",
			},
			ExpectFieldError: cli.ErrMissingField(cli.NamespaceFlagName),
		},
		{
			Name: "missing name",
			Options: &cli.ResourceOptions{
				Namespace: "default",
			},
			ExpectFieldError: cli.ErrMissingField(cli.NameArgumentName),
		},
	}

	table.Run(t)
}

func TestDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "default",
			Options: &cli.DeleteOptions{
				Namespace: "default",
			},
			ExpectFieldError: cli.ErrMissingOneOf(cli.AllFlagName, cli.NamesArgumentName),
		},
		{
			Name: "single name",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function"},
			},
			ShouldValidate: true,
		},
		{
			Name: "multiple names",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function", "my-other-function"},
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid name",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my.function"},
			},
			ExpectFieldError: cli.ErrInvalidValue("my.function", cli.CurrentField).ViaFieldIndex(cli.NamesArgumentName, 0),
		},
		{
			Name: "all",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				All:       true,
			},
			ShouldValidate: true,
		},
		{
			Name: "all with name",
			Options: &cli.DeleteOptions{
				Namespace: "default",
				Names:     []string{"my-function"},
				All:       true,
			},
			ExpectFieldError: cli.ErrMultipleOneOf(cli.AllFlagName, cli.NamesArgumentName),
		},
		{
			Name: "missing namespace",
			Options: &cli.DeleteOptions{
				Names: []string{"my-function"},
			},
			ExpectFieldError: cli.ErrMissingField(cli.NamespaceFlagName),
		},
	}

	table.Run(t)
}
