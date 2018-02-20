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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/projectriff/riff-cli/pkg/options"
	"strings"
	"github.com/projectriff/riff-cli/pkg/docker"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"os"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"path/filepath"
	"github.com/projectriff/riff-cli/cmd/utils"
	"github.com/projectriff/riff-cli/cmd/opts"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a function container",
	Long: `Build the function based on the code available in the path directory, using the name
  and version specified for the image that is built.`,
	Example: `  riff build -n <name> -v <version> -f <path> [--push]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return build(options.GetBuildOptions(opts.CreateOptions))
	},
	//TODO: DRY
	PreRun: func(cmd *cobra.Command, args []string) {
		if !opts.CreateOptions.Initialized {
			utils.MergeBuildOptions(*cmd.Flags(), &opts.CreateOptions)

			if len(args) > 0 {
				if len(args) == 1 && opts.CreateOptions.FilePath == "" {
					opts.CreateOptions.FilePath = args[0]
				} else {
					ioutils.Errorf("Invalid argument(s) %v\n", args)
					cmd.Usage()
					os.Exit(1)
				}
			}

			err := options.ValidateAndCleanInitOptions(&opts.CreateOptions.InitOptions)
			if err != nil {
				ioutils.Error(err)
				os.Exit(1)
			}
		}
		opts.CreateOptions.Initialized = true
	},
}

func build(opts options.BuildOptions) error {
	buildArgs := buildArgs(opts)
	pushArgs := pushArgs(opts)
	if opts.DryRun {
		fmt.Printf("\nBuild command: docker %s\n", strings.Join(buildArgs, " "))
		if (opts.Push) {
			fmt.Printf("\nPush command: docker %s\n", strings.Join(pushArgs, " "))
		}
		fmt.Println("")
		return nil
	}

	fmt.Println("Building image ...")
	out, err := docker.Exec(buildArgs)
	if err != nil {
		ioutils.Errorf("Error %v\n", err)
		return err
	}
	fmt.Println(out)

	if opts.Push {
		fmt.Println("Pushing image...")
		out, err = docker.Exec(pushArgs)
		if err != nil {
			ioutils.Errorf("Error %v\n", err)
			return err
		}
		fmt.Println(out)
	}

	return nil
}

func buildArgs(opts options.BuildOptions) []string {
	image := options.ImageName(opts)
	path := opts.FilePath
	if !osutils.IsDirectory(opts.FilePath) {
		path = filepath.Dir(path)
	}
	return []string{"build", "-t", image, path}
}

func pushArgs(opts options.BuildOptions) []string {
	image := options.ImageName(opts)
	return []string{"push", image}
}

func init() {
	rootCmd.AddCommand(buildCmd)
	utils.CreateBuildFlags(buildCmd.Flags())
}
