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
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a function container",
	Long: `Build the function based on the code available in the path directory, using the name
  and version specified for the image that is built.`,
	Example: `riff build -n <name> -v <version> -f <path> [--push]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return build(createOptions)
	},
	//TODO: DRY
	PreRun: func(cmd *cobra.Command, args []string) {
		if !createOptions.Initialized {
			createOptions = options.CreateOptions{}
			mergeBuildOptions(*cmd.Flags(), &createOptions)

			if len(args) > 0 {
				if len(args) == 1 && createOptions.FunctionPath == "" {
					createOptions.FunctionPath = args[0]
				} else {
					ioutils.Errorf("Invalid argument(s) %v\n", args)
					cmd.Usage()
					os.Exit(1)
				}
			}

			err := options.ValidateAndCleanInitOptions(&createOptions.InitOptions)
			if err != nil {
				ioutils.Error(err)
				os.Exit(1)
			}
		}
		createOptions.Initialized = true
	},
}

func build(opts options.CreateOptions) error {
	buildArgs := buildArgs(opts)
	pushArgs := pushArgs(opts)
	if opts.DryRun {
		fmt.Printf("\nBuild command: docker build %s\n", strings.Join(buildArgs, " "))
		if (opts.Push) {
			fmt.Printf("\nPush command: docker %s\n", strings.Join(pushArgs, " "))
		}
		return nil
	}

	out, err := docker.Exec(buildArgs)
	if err != nil {
		ioutils.Errorf("Error %v\n", err)
		return err
	}
	fmt.Println(out)

	if opts.Push {
		out, err = docker.Exec(pushArgs)
		if err != nil {
			ioutils.Errorf("Error %v\n", err)
			return err
		}
		fmt.Println(out)
	}

	return nil
}

func buildArgs(opts options.CreateOptions) []string {
	image := options.ImageName(opts.InitOptions)
	path := opts.FunctionPath
	if !osutils.IsDirectory(opts.FunctionPath) {
		path = filepath.Dir(path)
	}
	return []string{"build", "-t", image, path}
}

func pushArgs(opts options.CreateOptions) []string {
	image := options.ImageName(opts.InitOptions)
	return []string{"push", image}
}

func init() {
	rootCmd.AddCommand(buildCmd)
	createBuildFlags(buildCmd.Flags())
}
