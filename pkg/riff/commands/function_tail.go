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
	"time"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FunctionTailOptions struct {
	cli.ResourceOptions
	Since string
}

func (opts *FunctionTailOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	if opts.Since != "" {
		if _, err := time.ParseDuration(opts.Since); err != nil {
			errs = errs.Also(cli.ErrInvalidValue(opts.Since, cli.SinceFlagName))
		}
	}

	return errs
}

func (opts *FunctionTailOptions) Exec(ctx context.Context, c *cli.Config) error {
	function, err := c.Build().Functions(opts.Namespace).Get(opts.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	since := time.Second
	if opts.Since != "" {
		// error is protected by Validate()
		since, _ = time.ParseDuration(opts.Since)
	}
	return c.Kail.FunctionLogs(ctx, function, since, c.Stdout)
}

func NewFunctionTailCommand(c *cli.Config) *cobra.Command {
	opts := &FunctionTailOptions{}

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "watch build logs",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s function tail my-function", c.Name),
			fmt.Sprintf("%s function tail my-function %s 1h", c.Name, cli.SinceFlagName),
		}, "\n"),
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Since, cli.StripDash(cli.SinceFlagName), "", "time `duration` to start reading logs from")

	return cmd
}
