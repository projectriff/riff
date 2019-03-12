/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func Docs(rootCmd *cobra.Command, fs Filesystem) *cobra.Command {

	var directory string

	var docsCmd = &cobra.Command{
		Use:    "docs",
		Short:  "generate riff command documentation",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GenerateDocs(rootCmd, directory, fs)
		},
	}
	docsCmd.Flags().StringVarP(&directory, "dir", "d", "docs", "the output directory for the docs.")
	return docsCmd
}

func GenerateDocs(rootCommand *cobra.Command, directory string, fs Filesystem) error {
	if err := fs.MkdirAll(directory, 0744); err != nil {
		return err
	}
	return doc.GenMarkdownTree(rootCommand, directory)
}
