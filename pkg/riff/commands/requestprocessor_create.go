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

	"github.com/knative/pkg/apis"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/parsers"
	"github.com/projectriff/riff/pkg/validation"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RequestProcessorCreateOptions struct {
	Namespace string
	Name      string

	ItemName       string
	Image          string
	ApplicationRef string
	FunctionRef    string

	Env     []string
	EnvFrom []string
}

func (opts *RequestProcessorCreateOptions) Validate(ctx context.Context) *apis.FieldError {
	errs := &apis.FieldError{}

	if opts.Namespace == "" {
		errs = errs.Also(apis.ErrMissingField("namespace"))
	}

	if opts.Name == "" {
		errs = errs.Also(apis.ErrInvalidValue(opts.Name, "name"))
	} else {
		errs = errs.Also(validation.K8sName(opts.Name, "name"))
	}

	if opts.ItemName == "" {
		errs = errs.Also(apis.ErrMissingField("item"))
	} else {
		errs = errs.Also(validation.K8sName(opts.ItemName, "item"))
	}

	// application-ref, build-ref and image are mutually exclusive
	used := []string{}
	unused := []string{}

	if opts.ApplicationRef != "" {
		used = append(used, "application-ref")
	} else {
		unused = append(unused, "application-ref")
	}

	if opts.FunctionRef != "" {
		used = append(used, "function-ref")
	} else {
		unused = append(unused, "function-ref")
	}

	if opts.Image != "" {
		used = append(used, "image")
	} else {
		unused = append(unused, "image")
	}

	if len(used) == 0 {
		errs = errs.Also(apis.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(apis.ErrMultipleOneOf(used...))
	}

	errs = errs.Also(validation.EnvVars(opts.Env, "env"))

	// TODO validate EnvFrom

	return errs
}

func NewRequestProcessorCreateCommand(c *cli.Config) *cobra.Command {
	opts := &RequestProcessorCreateOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "<todo>",
		Example: "<todo>",
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE: func(cmd *cobra.Command, args []string) error {
			processor := &requestv1alpha1.RequestProcessor{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: opts.Namespace,
					Name:      opts.Name,
				},
				Spec: requestv1alpha1.RequestProcessorSpec{
					{
						Name: opts.ItemName,
						Template: &corev1.PodSpec{
							Containers: []corev1.Container{{}},
						},
					},
				},
			}

			if opts.ApplicationRef != "" {
				processor.Spec[0].Build = &requestv1alpha1.Build{
					ApplicationRef: opts.ApplicationRef,
				}
			}
			if opts.FunctionRef != "" {
				processor.Spec[0].Build = &requestv1alpha1.Build{
					FunctionRef: opts.FunctionRef,
				}
			}
			if opts.Image != "" {
				processor.Spec[0].Template.Containers[0].Image = opts.Image
			}

			for _, env := range opts.Env {
				if processor.Spec[0].Template.Containers[0].Env == nil {
					processor.Spec[0].Template.Containers[0].Env = []corev1.EnvVar{}
				}
				processor.Spec[0].Template.Containers[0].Env = append(processor.Spec[0].Template.Containers[0].Env, parsers.EnvVar(env))
			}
			for range opts.EnvFrom {
				return fmt.Errorf("not implemented")
			}

			processor, err := c.Request().RequestProcessors(opts.Namespace).Create(processor)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created request processor %q\n", processor.Name)
			return nil
		},
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.ItemName, "item", "", "<todo>")
	cmd.Flags().StringVar(&opts.Image, "image", "", "<todo>")
	cmd.Flags().StringVar(&opts.ApplicationRef, "application-ref", "", "<todo>")
	cmd.Flags().StringVar(&opts.FunctionRef, "function-ref", "", "<todo>")
	cmd.Flags().StringArrayVar(&opts.Env, "env", []string{}, "<todo>")
	cmd.Flags().StringArrayVar(&opts.EnvFrom, "env-from", []string{}, "<todo>")

	return cmd
}
