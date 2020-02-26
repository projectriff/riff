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
	"regexp"
	"strings"
	"time"

	"github.com/projectriff/cli/pkg/parsers"
	"github.com/projectriff/cli/pkg/validation"

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/cli/options"
	"github.com/projectriff/cli/pkg/k8s"
	"github.com/projectriff/cli/pkg/race"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProcessorCreateOptions struct {
	options.ResourceOptions

	Image        string
	ContainerRef string
	FunctionRef  string

	Env     []string
	EnvFrom []string

	Inputs  []string
	Outputs []string

	Tail        bool
	WaitTimeout string

	DryRun bool
}

var (
	_ cli.Validatable = (*ProcessorCreateOptions)(nil)
	_ cli.Executable  = (*ProcessorCreateOptions)(nil)
	_ cli.DryRunable  = (*ProcessorCreateOptions)(nil)
)

func (opts *ProcessorCreateOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	// build-ref and image are mutually exclusive
	used := []string{}
	unused := []string{}

	if opts.ContainerRef != "" {
		used = append(used, cli.ContainerRefFlagName)
	} else {
		unused = append(unused, cli.ContainerRefFlagName)
	}

	if opts.FunctionRef != "" {
		used = append(used, cli.FunctionRefFlagName)
	} else {
		unused = append(unused, cli.FunctionRefFlagName)
	}

	errs = errs.Also(validation.EnvVars(opts.Env, cli.EnvFlagName))
	errs = errs.Also(validation.EnvVarFroms(opts.EnvFrom, cli.EnvFromFlagName))

	if opts.Image != "" {
		used = append(used, cli.ImageFlagName)
	} else {
		unused = append(unused, cli.ImageFlagName)
	}

	if len(used) == 0 {
		errs = errs.Also(cli.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(cli.ErrMultipleOneOf(used...))
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
	var err error
	inputs, err := parseInputStreamBindings(opts.Inputs)
	if err != nil {
		return err
	}
	outputs, err := parseOutputStreamBindings(opts.Outputs)
	if err != nil {
		return err
	}
	processor := &streamingv1alpha1.Processor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: streamingv1alpha1.ProcessorSpec{
			Inputs:  inputs,
			Outputs: outputs,
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
		},
	}

	if opts.ContainerRef != "" {
		processor.Spec.Build = &streamingv1alpha1.Build{
			ContainerRef: opts.ContainerRef,
		}
	}
	if opts.FunctionRef != "" {
		processor.Spec.Build = &streamingv1alpha1.Build{
			FunctionRef: opts.FunctionRef,
		}
	}

	for _, env := range opts.Env {
		if processor.Spec.Template.Spec.Containers[0].Env == nil {
			processor.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{}
		}
		processor.Spec.Template.Spec.Containers[0].Env = append(processor.Spec.Template.Spec.Containers[0].Env, parsers.EnvVar(env))
	}
	for _, env := range opts.EnvFrom {
		if processor.Spec.Template.Spec.Containers[0].Env == nil {
			processor.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{}
		}
		processor.Spec.Template.Spec.Containers[0].Env = append(processor.Spec.Template.Spec.Containers[0].Env, parsers.EnvVarFrom(env))
	}

	if opts.Image != "" {
		processor.Spec.Template.Spec.Containers[0].Image = opts.Image
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, processor, processor.GetGroupVersionKind())
	} else {
		var err error
		processor, err = c.StreamingRuntime().Processors(opts.Namespace).Create(processor)
		if err != nil {
			return err
		}
	}
	c.Successf("Created processor %q\n", processor.Name)
	if opts.Tail {
		c.Infof("Waiting for processor %q to become ready...\n", processor.Name)
		// err guarded by Validate()
		timeout, _ := time.ParseDuration(opts.WaitTimeout)
		err := race.Run(ctx, timeout,
			func(ctx context.Context) error {
				return k8s.WaitUntilReady(ctx, c.StreamingRuntime().RESTClient(), "processors", processor)
			},
			func(ctx context.Context) error {
				return c.Kail.StreamingProcessorLogs(ctx, processor, cli.TailSinceCreateDefault, c.Stdout)
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
		c.Successf("Processor %q is ready\n", processor.Name)
	}
	return nil
}

func (opts *ProcessorCreateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewProcessorCreateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ProcessorCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a processor to apply a function to messages on streams",
		Long: strings.TrimSpace(`
Creates a processor within a namespace.

The processor is configured with a function or container reference and multiple
input and/or output streams.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s streaming processor create my-processor %s my-func %s my-input-stream", c.Name, cli.FunctionRefFlagName, cli.InputFlagName),
			fmt.Sprintf("%s streaming processor create my-processor %s my-func %s input:my-input-stream %s my-join-stream@earliest %s out:my-output-stream", c.Name, cli.FunctionRefFlagName, cli.InputFlagName, cli.InputFlagName, cli.OutputFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Image, cli.StripDash(cli.ImageFlagName), "", "container `image` to deploy")
	cmd.Flags().StringVar(&opts.ContainerRef, cli.StripDash(cli.ContainerRefFlagName), "", "`name` of container to deploy")
	_ = cmd.MarkFlagCustom(cli.StripDash(cli.ContainerRefFlagName), "__"+c.Name+"_list_containers")
	cmd.Flags().StringVar(&opts.FunctionRef, cli.StripDash(cli.FunctionRefFlagName), "", "`name` of function to deploy")
	_ = cmd.MarkFlagCustom(cli.StripDash(cli.FunctionRefFlagName), "__"+c.Name+"_list_functions")
	cmd.Flags().StringArrayVar(&opts.Inputs, cli.StripDash(cli.InputFlagName), []string{}, "`name` of stream to read messages from (or [<alias>:]<stream>[@<earliest|latest>], may be set multiple times)")
	cmd.Flags().StringArrayVar(&opts.Outputs, cli.StripDash(cli.OutputFlagName), []string{}, "`name` of stream to write messages to (or [<alias>:]<stream>, may be set multiple times)")
	cmd.Flags().BoolVar(&opts.Tail, cli.StripDash(cli.TailFlagName), false, "watch processor logs")
	cmd.Flags().StringVar(&opts.WaitTimeout, cli.StripDash(cli.WaitTimeoutFlagName), "10m", "`duration` to wait for the processor to become ready when watching logs")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")
	cmd.Flags().StringArrayVar(&opts.Env, cli.StripDash(cli.EnvFlagName), []string{}, fmt.Sprintf("environment `variable` defined as a key value pair separated by an equals sign, example %q (may be set multiple times)", fmt.Sprintf("%s MY_VAR=my-value", cli.EnvFlagName)))
	cmd.Flags().StringArrayVar(&opts.EnvFrom, cli.StripDash(cli.EnvFromFlagName), []string{}, fmt.Sprintf("environment `variable` from a config map or secret, example %q, %q (may be set multiple times)", fmt.Sprintf("%s MY_SECRET_VALUE=secretKeyRef:my-secret-name:key-in-secret", cli.EnvFromFlagName), fmt.Sprintf("%s MY_CONFIG_MAP_VALUE=configMapKeyRef:my-config-map-name:key-in-config-map", cli.EnvFromFlagName)))

	return cmd
}

// Parse input stream bindings. Valid values are of the form [<alias>:]<stream>[@<offset>].
// Default values are handled on the server side.
func parseInputStreamBindings(raw []string) ([]streamingv1alpha1.InputStreamBinding, error) {
	bindings := make([]streamingv1alpha1.InputStreamBinding, len(raw))
	pattern := regexp.MustCompile(`^(?:([^:]+):)?([^:@]+)(?:@(earliest|latest))?$`)
	for i, s := range raw {
		parts := pattern.FindStringSubmatch(s)
		if len(parts) != 4 {
			return nil, fmt.Errorf("malformed input stream reference %q, should be of the form [<alias>:]<stream>[@<offset>]", s)
		} else {
			bindings[i].Alias = parts[1]       // group 1
			bindings[i].Stream = parts[2]      // group 2
			bindings[i].StartOffset = parts[3] // group 3
		}
	}
	return bindings, nil
}

// Parse output stream bindings. Valid values are of the form [<alias>:]<stream>.
// Default values are handled on the server side.
func parseOutputStreamBindings(raw []string) ([]streamingv1alpha1.OutputStreamBinding, error) {
	bindings := make([]streamingv1alpha1.OutputStreamBinding, len(raw))
	pattern := regexp.MustCompile(`^(?:([^:]+):)?([^:]+)$`)
	for i, s := range raw {
		parts := pattern.FindStringSubmatch(s)
		if len(parts) != 3 {
			return nil, fmt.Errorf("malformed output stream reference %q, should be of the form [<alias>:]<stream>", s)
		} else {
			bindings[i].Alias = parts[1]  // group 1
			bindings[i].Stream = parts[2] // group 2
		}
	}
	return bindings, nil
}
