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

package commands

import (
	"context"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
)

type CompletionOptions struct {
	Shell string
	cmd   *cobra.Command
}

func (o *CompletionOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}
	if o.Shell == "" {
		errs = errs.Also(cli.ErrMissingField("shell"))
	} else if o.Shell != "bash" && o.Shell != "zsh" {
		errs = errs.Also(cli.ErrInvalidValue(o.Shell, "shell"))
	}
	return errs
}

func (o *CompletionOptions) Exec(ctx context.Context, c *cli.Config) error {
	switch o.Shell {
	case "bash":
		return o.cmd.Root().GenBashCompletion(o.cmd.OutOrStdout())
	case "zsh":
		return o.cmd.Root().GenZshCompletion(o.cmd.OutOrStdout())
	default:
		panic("invalid shell: " + o.Shell) // protected by o.Validate()
	}
	return nil
}

func NewCompletionCommand(c *cli.Config) *cobra.Command {

	opts := &CompletionOptions{}

	cmd := &cobra.Command{
		Use:     "completion",
		Short:   "<todo>",
		Example: "<todo>",
		Args:    cli.Args(cli.NameArg(&opts.Shell)),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}
	opts.cmd = cmd

	return cmd
}
