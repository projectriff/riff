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
	"fmt"
	"strings"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
)

type CompletionOptions struct {
	Shell string
}

var (
	_ cli.Validatable = (*CompletionOptions)(nil)
	_ cli.Executable  = (*CompletionOptions)(nil)
)

func (opts *CompletionOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

	if opts.Shell == "" {
		errs = errs.Also(cli.ErrMissingField(cli.ShellFlagName))
	} else if opts.Shell != "bash" && opts.Shell != "zsh" {
		errs = errs.Also(cli.ErrInvalidValue(opts.Shell, cli.ShellFlagName))
	}

	return errs
}

func (opts *CompletionOptions) Exec(ctx context.Context, c *cli.Config) error {
	cmd := cli.CommandFromContext(ctx)
	switch opts.Shell {
	case "bash":
		return cmd.Root().GenBashCompletion(c.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(c.Stdout)
	}
	// protected by opts.Validate()
	panic("invalid shell: " + opts.Shell)
}

func NewCompletionCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &CompletionOptions{}

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "generate shell completion script",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s completion", c.Name),
			fmt.Sprintf("%s completion %s zsh", c.Name, cli.ShellFlagName),
		}, "\n"),
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cmd.Flags().StringVar(&opts.Shell, cli.StripDash(cli.ShellFlagName), "bash", "`shell` to generate completion for: bash or zsh")

	return cmd
}
