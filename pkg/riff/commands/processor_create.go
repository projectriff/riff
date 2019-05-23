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
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProcessorCreateOptions struct {
	cli.ResourceOptions

	FunctionRef string
	Inputs      []string
	Outputs     []string
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
	return nil
}

func NewProcessorCreateCommand(c *cli.Config) *cobra.Command {
	opts := &ProcessorCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "process messages with a function",
		Long: `
<todo>
`,
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
	cmd.Flags().StringVar(&opts.FunctionRef, cli.StripDash(cli.FunctionRefFlagName), "", "function build to deploy")
	cmd.Flags().StringArrayVar(&opts.Inputs, cli.StripDash(cli.InputFlagName), []string{}, "stream to read messages from")
	cmd.Flags().StringArrayVar(&opts.Outputs, cli.StripDash(cli.OutputFlagName), []string{}, "stream to write messages to")

	return cmd
}
