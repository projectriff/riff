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
	"github.com/projectriff/system/pkg/apis/build"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialDeleteOptions struct {
	cli.DeleteOptions
}

func (opts *CredentialDeleteOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	errs = errs.Also(opts.DeleteOptions.Validate(ctx))

	return errs
}

func NewCredentialDeleteCommand(c *cli.Config) *cobra.Command {
	opts := &CredentialDeleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "<todo>",
		Example: "<todo>",
		Args: cli.Args(
			cli.NamesArg(&opts.Names),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := c.Core().Secrets(opts.Namespace)

			if opts.All {
				return client.DeleteCollection(nil, metav1.ListOptions{
					LabelSelector: build.CredentialLabelKey,
				})
			}

			for _, name := range opts.Names {
				// TODO check for the matching label before deleting
				if err := client.Delete(name, nil); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().BoolVar(&opts.All, cli.StripDash(cli.AllFlagName), false, "<todo>")

	return cmd
}
