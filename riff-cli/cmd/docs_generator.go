/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra/doc"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"os"
)

var directory string
var createDir bool

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "generate riff-cli command documentation",
	Long:  `Generate riff-cli command documentation`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {

		if !osutils.IsDirectory(directory) {
			os.Mkdir(directory,0744)
		}

		err := doc.GenMarkdownTree(rootCmd, directory)
		if err != nil {
			ioutils.Errorf("Doc generation failed %v\n", err)
			os.Exit(1)
		}
	},
}


func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringVarP(&directory, "dir", "d", osutils.Path(osutils.GetCWD()+"/docs"),"the output directory for the docs.")

}