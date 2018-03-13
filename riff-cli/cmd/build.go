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

	"errors"
	"path/filepath"
	"strings"

	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/docker"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/pkg/functions"
)

type BuildOptions struct {
	FilePath     string
	FunctionName string
	Version      string
	RiffVersion  string
	UserAccount  string
	Push         bool
	DryRun       bool
}

func (this BuildOptions) GetFunctionName() string {
	return this.FunctionName
}

func (this BuildOptions) GetVersion() string {
	return this.Version
}

func (this BuildOptions) GetUserAccount() string {
	return this.UserAccount
}

func Build() (*cobra.Command, *BuildOptions) {

	buildOptions := BuildOptions{}

	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build a function container",
		Long: `Build the function based on the code available in the path directory, using the name
and version specified for the image that is built.`,
		Example: `  riff build -n <name> -v <version> -f <path> [--push]`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := build(buildOptions)
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			//TODO: DRY
			if len(args) > 0 {
				if len(args) == 1 && buildOptions.FilePath == "" {
					buildOptions.FilePath = args[0]
				} else {
					return errors.New(fmt.Sprintf("Invalid argument(s) %v\n", args))
				}
			}

			err := validateOptions(&buildOptions)
			if err != nil {
				return err
			}
			return nil
		},
	}

	buildCmd.Flags().BoolVar(&buildOptions.DryRun, "dry-run", false, "print generated function artifacts content to stdout only")
	buildCmd.Flags().StringVarP(&buildOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	buildCmd.Flags().StringVarP(&buildOptions.FunctionName, "name", "n", "", "the name of the function (defaults to the name of the current directory)")
	buildCmd.Flags().BoolVar(&buildOptions.Push, "push", false, "push the image to Docker registry")
	buildCmd.Flags().StringVar(&buildOptions.RiffVersion, "riff-version", utils.DefaultValues.RiffVersion, "the version of riff to use when building containers")
	buildCmd.Flags().StringVarP(&buildOptions.Version, "version", "v", utils.DefaultValues.Version, "the version of the function image")
	buildCmd.Flags().StringVarP(&buildOptions.UserAccount, "useraccount", "u", utils.DefaultValues.UserAccount, "the Docker user account to be used for the image repository")

	return buildCmd, &buildOptions
}

func build(opts BuildOptions) error {
	buildArgs := buildArgs(opts)
	pushArgs := pushArgs(opts)
	if opts.DryRun {
		fmt.Printf("\nBuild command: docker %s\n", strings.Join(buildArgs, " "))
		if opts.Push {
			fmt.Printf("\nPush command: docker %s\n", strings.Join(pushArgs, " "))
		}
		fmt.Println("")
		return nil
	}

	fmt.Println("Building image ...")
	docker.Exec(buildArgs)

	if opts.Push {
		fmt.Println("Pushing image...")
		docker.Exec(pushArgs)
	}

	return nil
}

func buildArgs(opts BuildOptions) []string {
	image := options.ImageName(opts)
	path := opts.FilePath
	if !osutils.IsDirectory(opts.FilePath) {
		path = filepath.Dir(path)
	}
	return []string{"build", "-t", image, path}
}

func pushArgs(opts BuildOptions) []string {
	image := options.ImageName(opts)
	return []string{"push", image}
}

func validateOptions(options *BuildOptions) error {
	options.FilePath = filepath.Clean(options.FilePath)

	if options.FilePath == "" {
		path, _ := filepath.Abs(".")
		options.FilePath = path
	}

	var err error
	if options.FunctionName == "" {
		options.FunctionName, err = functions.FunctionNameFromPath(options.FilePath)
		if err != nil {
			return err
		}
	}
	return nil
}
