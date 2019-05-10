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

	"github.com/knative/pkg/apis"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/validation"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FunctionDeleteOptions struct {
	Namespace string
	Names     []string
	All       bool
}

func (opts *FunctionDeleteOptions) Validate(ctx context.Context) *apis.FieldError {
	errs := &apis.FieldError{}

	if opts.Namespace == "" {
		errs = errs.Also(apis.ErrMissingField("namespace"))
	}

	if opts.All && len(opts.Names) != 0 {
		errs = errs.Also(apis.ErrMultipleOneOf("all", "names"))
	}
	if !opts.All && len(opts.Names) == 0 {
		errs = errs.Also(apis.ErrMissingOneOf("all", "names"))
	}

	errs = errs.Also(validation.K8sNames(opts.Names, "names"))

	return errs
}

func NewFunctionDeleteCommand(c *cli.Config) *cobra.Command {
	opts := &FunctionDeleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "<todo>",
		Example: "<todo>",
		Args: cli.Args(
			cli.NamesArg(&opts.Names),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := c.Build().Functions(opts.Namespace)

			if opts.All {
				return client.DeleteCollection(nil, metav1.ListOptions{})
			}

			for _, name := range opts.Names {
				if err := client.Delete(name, nil); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().BoolVar(&opts.All, "all", false, "<todo>")

	return cmd
}
