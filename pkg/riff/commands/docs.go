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
	"os"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type DocsOptions struct {
	Directory string
}

func (opts *DocsOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	if opts.Directory == "" {
		errs = errs.Also(cli.ErrMissingField(cli.DirectoryFlagName))
	}

	return errs
}

func NewDocsCommand(c *cli.Config) *cobra.Command {
	opts := &DocsOptions{}

	cmd := &cobra.Command{
		Use:     "docs",
		Short:   "generate docs in Markdown for this CLI",
		Example: fmt.Sprintf("%s docs", c.Name),
		Hidden:  true,
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(opts),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(opts.Directory, 0744); err != nil {
				return err
			}
			root := cmd.Root()
			if noColorFlag := root.Flag(cli.StripDash(cli.NoColorFlagName)); noColorFlag != nil {
				// force default to false for doc generation no matter the environment
				noColorFlag.DefValue = "false"
			}

			return doc.GenMarkdownTree(root, opts.Directory)
		},
	}

	cmd.Flags().StringVarP(&opts.Directory, cli.StripDash(cli.DirectoryFlagName), "d", "docs", "the output directory for the docs")

	return cmd
}
