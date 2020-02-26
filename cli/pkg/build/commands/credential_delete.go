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

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/cli/options"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialDeleteOptions struct {
	options.DeleteOptions
}

var (
	_ cli.Validatable = (*CredentialDeleteOptions)(nil)
	_ cli.Executable  = (*CredentialDeleteOptions)(nil)
)

func (opts *CredentialDeleteOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.DeleteOptions.Validate(ctx))

	return errs
}

func (opts *CredentialDeleteOptions) Exec(ctx context.Context, c *cli.Config) error {
	client := c.Core().Secrets(opts.Namespace)

	if opts.All {
		err := client.DeleteCollection(nil, metav1.ListOptions{
			LabelSelector: buildv1alpha1.CredentialLabelKey,
		})
		if err != nil {
			return err
		}
		c.Successf("Deleted credentials in namespace %q\n", opts.Namespace)
		return nil
	}

	for _, name := range opts.Names {
		// TODO check for the matching label before deleting
		if err := client.Delete(name, nil); err != nil {
			return err
		}
		c.Successf("Deleted credential %q\n", name)
	}

	return nil
}

func NewCredentialDeleteCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &CredentialDeleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete credential(s)",
		Long: strings.TrimSpace(`
Delete one or more credentials by name or all credentials within a namespace.

Deleting a credential will cause builds that depend on the credential to fail
unless another credential for the same registry is available.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s credential delete my-creds", c.Name),
			fmt.Sprintf("%s credential delete %s ", c.Name, cli.AllFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NamesArg(&opts.Names),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().BoolVar(&opts.All, cli.StripDash(cli.AllFlagName), false, "delete all credentials within the namespace")

	return cmd
}
