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
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StreamDeleteOptions struct {
	cli.DeleteOptions
}

func (opts *StreamDeleteOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	errs = errs.Also(opts.DeleteOptions.Validate(ctx))

	return errs
}

func (opts *StreamDeleteOptions) Exec(ctx context.Context, c *cli.Config) error {
	client := c.Stream().Streams(opts.Namespace)

	if opts.All {
		if err := client.DeleteCollection(nil, metav1.ListOptions{}); err != nil {
			return err
		}
		c.Successf("Deleted streams in namespace %q\n", opts.Namespace)
		return nil
	}

	for _, name := range opts.Names {
		if err := client.Delete(name, nil); err != nil {
			return err
		}
		c.Successf("Deleted stream %q\n", name)
	}

	return nil
}

func NewStreamDeleteCommand(c *cli.Config) *cobra.Command {
	opts := &StreamDeleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete stream(s)",
		Long: `
<todo>
`,
		Example: strings.Join([]string{
			fmt.Sprintf("%s stream delete my-stream", c.Name),
			fmt.Sprintf("%s stream delete %s ", c.Name, cli.AllFlagName),
		}, "\n"),
		Args: cli.Args(
			cli.NamesArg(&opts.Names),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().BoolVar(&opts.All, cli.StripDash(cli.AllFlagName), false, "delete all streams within the namespace")

	return cmd
}
