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
	"path"
	"path/filepath"
	"strings"

	"github.com/projectriff/cli/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type DocsOptions struct {
	Directory string
}

var (
	_ cli.Validatable = (*DocsOptions)(nil)
)

func (opts *DocsOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	if opts.Directory == "" {
		errs = errs.Also(cli.ErrMissingField(cli.DirectoryFlagName))
	}

	return errs
}

func NewDocsCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &DocsOptions{}

	cmd := &cobra.Command{
		Use:     "docs",
		Short:   "generate docs in Markdown for this CLI",
		Example: fmt.Sprintf("%s docs", c.Name),
		Hidden:  true,
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(opts.Directory, 0744); err != nil {
				return err
			}
			root := cmd.Root()
			if noColorFlag := root.Flag(cli.StripDash(cli.NoColorFlagName)); noColorFlag != nil {
				// force default to false for doc generation no matter the environment
				noColorFlag.DefValue = "false"
			}

			// hack to rewrite the CommandPath content to add args
			cli.Visit(root, func(cmd *cobra.Command) error {
				if !cmd.HasSubCommands() {
					cmd.Use = cmd.Use + cli.FormatArgs(cmd)
				}
				return nil
			})

			return doc.GenMarkdownTreeCustom(root, opts.Directory,
				func(filename string) string {
					name := filepath.Base(filename)
					base := strings.TrimSuffix(name, path.Ext(name))
					id := strings.Replace(base, "_", "-", -1)
					title := strings.Replace(base, "_", " ", -1)

					// frontmatter for docusaurus markdown
					// per https://docusaurus.io/docs/en/doc-markdown#documents
					fmTemplate := `---
id: %s
title: "%s"
---
`

					return fmt.Sprintf(fmTemplate, id, title)
				},
				func(name string) string {
					return name
				},
			)
		},
	}

	cmd.Flags().StringVarP(&opts.Directory, cli.StripDash(cli.DirectoryFlagName), "d", "docs", "the output `directory` for the docs")

	return cmd
}
