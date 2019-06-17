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
	"github.com/projectriff/riff/pkg/k8s"
	"github.com/projectriff/riff/pkg/race"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProcessorCreateOptions struct {
	cli.ResourceOptions

	FunctionRef string
	Inputs      []string
	Outputs     []string

	Tail        bool
	WaitTimeout string

	DryRun bool
}

var (
	_ cli.Validatable = (*ProcessorCreateOptions)(nil)
	_ cli.Executable  = (*ProcessorCreateOptions)(nil)
	_ cli.DryRunable  = (*ProcessorCreateOptions)(nil)
)

func (opts *ProcessorCreateOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

	errs = errs.Also(opts.ResourceOptions.Validate((ctx)))

	if opts.FunctionRef == "" {
		errs = errs.Also(cli.ErrMissingField(cli.FunctionRefFlagName))
	}

	if len(opts.Inputs) == 0 {
		errs = errs.Also(cli.ErrMissingField(cli.InputFlagName))
	}

	if opts.Tail {
		if opts.WaitTimeout == "" {
			errs = errs.Also(cli.ErrMissingField(cli.WaitTimeoutFlagName))
		} else if _, err := time.ParseDuration(opts.WaitTimeout); err != nil {
			errs = errs.Also(cli.ErrInvalidValue(opts.WaitTimeout, cli.WaitTimeoutFlagName))
		}
	}

	if opts.DryRun && opts.Tail {
		errs = errs.Also(cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName))
	}

	return errs
}

func (opts *ProcessorCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	processor := &streamv1alpha1.Processor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: streamv1alpha1.ProcessorSpec{
			FunctionRef: opts.FunctionRef,
			Inputs:      opts.Inputs,
			Outputs:     opts.Outputs,
		},
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, processor, processor.GetGroupVersionKind())
	} else {
		var err error
		processor, err = c.Stream().Processors(opts.Namespace).Create(processor)
		if err != nil {
			return err
		}
	}
	c.Successf("Created processor %q\n", processor.Name)
	if opts.Tail {
		// err guarded by Validate()
		timeout, _ := time.ParseDuration(opts.WaitTimeout)
		err := race.Run(ctx, timeout,
			func(ctx context.Context) error {
				return k8s.WaitUntilReady(ctx, c.Stream().RESTClient(), "processors", processor)
			},
			func(ctx context.Context) error {
				return c.Kail.ProcessorLogs(ctx, processor, cli.TailSinceCreateDefault, c.Stdout)
			},
		)
		if err == context.DeadlineExceeded {
			c.Errorf("Timeout after %q waiting for %q to become ready\n", opts.WaitTimeout, opts.Name)
			c.Infof("To view status run: %s processor list %s %s\n", c.Name, cli.NamespaceFlagName, opts.Namespace)
			c.Infof("To continue watching logs run: %s processor tail %s %s %s\n", c.Name, opts.Name, cli.NamespaceFlagName, opts.Namespace)
			err = cli.SilenceError(err)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (opts *ProcessorCreateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewProcessorCreateCommand(c *cli.Config) *cobra.Command {
	opts := &ProcessorCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a processor to apply a function to messages on streams",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s processor create my-processor %s my-func %s my-input-stream", c.Name, cli.FunctionRefFlagName, cli.InputFlagName),
			fmt.Sprintf("%s processor create my-processor %s my-func %s my-input-stream %s my-join-stream %s my-output-stream", c.Name, cli.FunctionRefFlagName, cli.InputFlagName, cli.InputFlagName, cli.OutputFlagName),
		}, "\n"),
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.FunctionRef, cli.StripDash(cli.FunctionRefFlagName), "", "`name` of function build to deploy")
	cmd.Flags().StringArrayVar(&opts.Inputs, cli.StripDash(cli.InputFlagName), []string{}, "`name` of stream to read messages from (may be set multiple times)")
	cmd.Flags().StringArrayVar(&opts.Outputs, cli.StripDash(cli.OutputFlagName), []string{}, "`name` of stream to write messages to (may be set multiple times)")
	cmd.Flags().BoolVar(&opts.Tail, cli.StripDash(cli.TailFlagName), false, "watch processor logs")
	cmd.Flags().StringVar(&opts.WaitTimeout, cli.StripDash(cli.WaitTimeoutFlagName), "10m", "`duration` to wait for the processor to become ready when watching logs")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")

	return cmd
}
