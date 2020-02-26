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

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/cli/options"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProcessorTailOptions struct {
	options.ResourceOptions

	Since string
}

var (
	_ cli.Validatable = (*ProcessorTailOptions)(nil)
	_ cli.Executable  = (*ProcessorTailOptions)(nil)
)

func (opts *ProcessorTailOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	if opts.Since != "" {
		if _, err := time.ParseDuration(opts.Since); err != nil {
			errs = errs.Also(cli.ErrInvalidValue(opts.Since, cli.SinceFlagName))
		}
	}

	return errs
}

func (opts *ProcessorTailOptions) Exec(ctx context.Context, c *cli.Config) error {
	processor, err := c.StreamingRuntime().Processors(opts.Namespace).Get(opts.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	since := cli.TailSinceDefault
	if opts.Since != "" {
		// error is protected by Validate()
		since, _ = time.ParseDuration(opts.Since)
	}
	return c.Kail.StreamingProcessorLogs(ctx, processor, since, c.Stdout)
}

func NewProcessorTailCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ProcessorTailOptions{}

	cmd := &cobra.Command{
		Use:   "tail",
		Short: "watch processor logs",
		Long: strings.TrimSpace(`
Stream runtime logs for a processor until canceled. To cancel, press Ctl-c in
the shell or kill the process.

As new processor pods are started, the logs are displayed. To show historical
logs use ` + cli.SinceFlagName + `.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s streaming processor tail my-processor", c.Name),
			fmt.Sprintf("%s streaming processor tail my-processor %s 1h", c.Name, cli.SinceFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Since, cli.StripDash(cli.SinceFlagName), "", "time `duration` to start reading logs from")
	cmd.Flag(cli.StripDash(cli.SinceFlagName)).Hidden = true

	return cmd
}
