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
	"github.com/projectriff/riff/pkg/parsers"
	"github.com/projectriff/riff/pkg/validation"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HandlerCreateOptions struct {
	cli.ResourceOptions

	Image          string
	ApplicationRef string
	FunctionRef    string

	Env     []string
	EnvFrom []string
}

func (opts *HandlerCreateOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	errs = errs.Also(opts.ResourceOptions.Validate((ctx)))

	// application-ref, build-ref and image are mutually exclusive
	used := []string{}
	unused := []string{}

	if opts.ApplicationRef != "" {
		used = append(used, cli.ApplicationRefFlagName)
	} else {
		unused = append(unused, cli.ApplicationRefFlagName)
	}

	if opts.FunctionRef != "" {
		used = append(used, cli.FunctionRefFlagName)
	} else {
		unused = append(unused, cli.FunctionRefFlagName)
	}

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

	errs = errs.Also(validation.EnvVars(opts.Env, cli.EnvFlagName))
	errs = errs.Also(validation.EnvVarFroms(opts.EnvFrom, cli.EnvFromFlagName))

	return errs
}

func (opts *HandlerCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	processor := &requestv1alpha1.Handler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: requestv1alpha1.HandlerSpec{
			Template: &corev1.PodSpec{
				Containers: []corev1.Container{{}},
			},
		},
	}

	if opts.ApplicationRef != "" {
		processor.Spec.Build = &requestv1alpha1.Build{
			ApplicationRef: opts.ApplicationRef,
		}
	}
	if opts.FunctionRef != "" {
		processor.Spec.Build = &requestv1alpha1.Build{
			FunctionRef: opts.FunctionRef,
		}
	}
	if opts.Image != "" {
		processor.Spec.Template.Containers[0].Image = opts.Image
	}

	for _, env := range opts.Env {
		if processor.Spec.Template.Containers[0].Env == nil {
			processor.Spec.Template.Containers[0].Env = []corev1.EnvVar{}
		}
		processor.Spec.Template.Containers[0].Env = append(processor.Spec.Template.Containers[0].Env, parsers.EnvVar(env))
	}
	for _, env := range opts.EnvFrom {
		if processor.Spec.Template.Containers[0].Env == nil {
			processor.Spec.Template.Containers[0].Env = []corev1.EnvVar{}
		}
		processor.Spec.Template.Containers[0].Env = append(processor.Spec.Template.Containers[0].Env, parsers.EnvVarFrom(env))
	}

	processor, err := c.Request().Handlers(opts.Namespace).Create(processor)
	if err != nil {
		return err
	}
	c.Successf("Created handler %q\n", processor.Name)
	return nil
}

func NewHandlerCreateCommand(c *cli.Config) *cobra.Command {
	opts := &HandlerCreateOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "<todo>",
		Example: "<todo>",
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Image, cli.StripDash(cli.ImageFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.ApplicationRef, cli.StripDash(cli.ApplicationRefFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.FunctionRef, cli.StripDash(cli.FunctionRefFlagName), "", "<todo>")
	cmd.Flags().StringArrayVar(&opts.Env, cli.StripDash(cli.EnvFlagName), []string{}, "<todo>")
	cmd.Flags().StringArrayVar(&opts.EnvFrom, cli.StripDash(cli.EnvFromFlagName), []string{}, "<todo>")

	return cmd
}
