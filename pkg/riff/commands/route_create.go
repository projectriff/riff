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

type RouteCreateOptions struct {
	cli.ResourceOptions

	Image          string
	ApplicationRef string
	FunctionRef    string

	Env []string
	// TODO implement
	// EnvFrom []string
}

func (opts *RouteCreateOptions) Validate(ctx context.Context) *cli.FieldError {
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

	return errs
}

func (opts *RouteCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	route := &requestv1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: requestv1alpha1.RouteSpec{
			Template: &corev1.PodSpec{
				Containers: []corev1.Container{{}},
			},
		},
	}

	if opts.ApplicationRef != "" {
		route.Spec.Build = &requestv1alpha1.Build{
			ApplicationRef: opts.ApplicationRef,
		}
	}
	if opts.FunctionRef != "" {
		route.Spec.Build = &requestv1alpha1.Build{
			FunctionRef: opts.FunctionRef,
		}
	}
	if opts.Image != "" {
		route.Spec.Template.Containers[0].Image = opts.Image
	}

	for _, env := range opts.Env {
		if route.Spec.Template.Containers[0].Env == nil {
			route.Spec.Template.Containers[0].Env = []corev1.EnvVar{}
		}
		route.Spec.Template.Containers[0].Env = append(route.Spec.Template.Containers[0].Env, parsers.EnvVar(env))
	}

	route, err := c.Request().Routes(opts.Namespace).Create(route)
	if err != nil {
		return err
	}
	c.Successf("Created route %q\n", route.Name)
	return nil
}

func NewRouteCreateCommand(c *cli.Config) *cobra.Command {
	opts := &RouteCreateOptions{}

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

	return cmd
}
