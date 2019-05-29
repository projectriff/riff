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
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProcessorCreateOptions struct {
	cli.ResourceOptions

	FunctionRef string
	Inputs      []string
	Outputs     []string

	Tail bool
}

func (opts *ProcessorCreateOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	errs = errs.Also(opts.ResourceOptions.Validate((ctx)))

	if opts.FunctionRef == "" {
		errs = errs.Also(cli.ErrMissingField(cli.FunctionRefFlagName))
	}

	if len(opts.Inputs) == 0 {
		errs = errs.Also(cli.ErrMissingField(cli.InputFlagName))
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

	processor, err := c.Stream().Processors(opts.Namespace).Create(processor)
	if err != nil {
		return err
	}
	c.Successf("Created processor %q\n", processor.Name)
	if opts.Tail {
		// cancel ctx when processor becomes ready
		ctx, cancel := context.WithCancel(ctx)
		go func() {
			defer cancel()
			processorWatch, err := c.Stream().Processors(opts.Namespace).Watch(metav1.ListOptions{
				ResourceVersion: processor.ResourceVersion,
			})
			if err != nil {
				c.Errorf("Error: %s\n", err)
				return
			}
			defer processorWatch.Stop()
			if err := k8s.WaitUntilReady(processor, processorWatch); err != nil {
				c.Errorf("Error: %s\n", err)
				return
			}
		}()
		return c.Kail.ProcessorLogs(ctx, processor, time.Minute, c.Stdout)
	}
	return nil
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

	return cmd
}
