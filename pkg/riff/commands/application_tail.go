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

type ApplicationTailOptions struct {
	cli.ResourceOptions
	Since string
}

var (
	_ cli.Validatable = (*ApplicationTailOptions)(nil)
	_ cli.Executable  = (*ApplicationTailOptions)(nil)
)

func (opts *ApplicationTailOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	if opts.Since != "" {
		if _, err := time.ParseDuration(opts.Since); err != nil {
			errs = errs.Also(cli.ErrInvalidValue(opts.Since, cli.SinceFlagName))
		}
	}

	return errs
}

func (opts *ApplicationTailOptions) Exec(ctx context.Context, c *cli.Config) error {
	application, err := c.Build().Applications(opts.Namespace).Get(opts.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	since := cli.TailSinceDefault
	if opts.Since != "" {
		// error is protected by Validate()
		since, _ = time.ParseDuration(opts.Since)
	}
	return c.Kail.ApplicationLogs(ctx, application, since, c.Stdout)
}

func NewApplicationTailCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ApplicationTailOptions{}

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "watch build logs",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s application tail my-application", c.Name),
			fmt.Sprintf("%s application tail my-application %s 1h", c.Name, cli.SinceFlagName),
		}, "\n"),
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Since, cli.StripDash(cli.SinceFlagName), "", "time `duration` to start reading logs from")

	return cmd
}
