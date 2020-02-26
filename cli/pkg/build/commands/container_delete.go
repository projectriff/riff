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

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/cli/options"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ContainerDeleteOptions struct {
	options.DeleteOptions
}

var (
	_ cli.Validatable = (*ContainerDeleteOptions)(nil)
	_ cli.Executable  = (*ContainerDeleteOptions)(nil)
)

func (opts *ContainerDeleteOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.DeleteOptions.Validate(ctx))

	return errs
}

func (opts *ContainerDeleteOptions) Exec(ctx context.Context, c *cli.Config) error {
	client := c.Build().Containers(opts.Namespace)

	if opts.All {
		if err := client.DeleteCollection(nil, metav1.ListOptions{}); err != nil {
			return err
		}
		c.Successf("Deleted containers in namespace %q\n", opts.Namespace)
		return nil
	}

	for _, name := range opts.Names {
		if err := client.Delete(name, nil); err != nil {
			return err
		}
		c.Successf("Deleted container %q\n", name)
	}

	return nil
}

func NewContainerDeleteCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ContainerDeleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete container(s)",
		Long: strings.TrimSpace(`
Delete one or more containers by name or all containers within a namespace.

Deleting a container prevents resolution of new images.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s container delete my-container", c.Name),
			fmt.Sprintf("%s container delete %s", c.Name, cli.AllFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NamesArg(&opts.Names),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().BoolVar(&opts.All, cli.StripDash(cli.AllFlagName), false, "delete all containers within the namespace")

	return cmd
}
