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

	"path/filepath"

	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/docker"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
)

type BuildOptions struct {
	FilePath     string
	FunctionName string
	Version      string
	UserAccount  string
	Push         bool
	DryRun       bool
}

func (bo BuildOptions) GetFunctionName() string {
	return bo.FunctionName
}

func (bo BuildOptions) GetVersion() string {
	return bo.Version
}

func (bo BuildOptions) GetUserAccount() string {
	return bo.UserAccount
}

func Build(realDocker docker.Docker, dryRunDocker docker.Docker) (*cobra.Command, *BuildOptions) {

	buildOptions := BuildOptions{}

	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build a function container image",
		Long: `Build the function based on the code available in the path directory, using the name
and version specified for the image that is built.`,
		Example: `  riff build -n <name> -v <version> -f <path> [--push]`,
		Args: utils.AliasFlagToSoleArg("filepath"),
		RunE: func(cmd *cobra.Command, args []string) error {
			dockerClient := realDocker
			if buildOptions.DryRun {
				dockerClient = dryRunDocker
			}
			err := build(buildOptions, dockerClient)
			if err != nil {
				cmd.SilenceUsage = true
			}
			return err
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			buildOptions.UserAccount = utils.GetUseraccountWithOverride("useraccount", *cmd.Flags())
			return validateBuildOptions(&buildOptions)
		},
	}

	buildCmd.Flags().BoolVar(&buildOptions.DryRun, "dry-run", false, "print generated function artifacts content to stdout only")
	buildCmd.Flags().StringVarP(&buildOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	buildCmd.Flags().StringVarP(&buildOptions.FunctionName, "name", "n", "", "the name of the function (defaults to the name of the current directory)")
	buildCmd.Flags().BoolVar(&buildOptions.Push, "push", false, "push the image to Docker registry")
	buildCmd.Flags().StringVarP(&buildOptions.Version, "version", "v", utils.DefaultValues.Version, "the version of the function image")
	buildCmd.Flags().StringVarP(&buildOptions.UserAccount, "useraccount", "u", utils.DefaultValues.UserAccount, "the Docker user account to be used for the image repository")
	return buildCmd, &buildOptions
}

func build(opts BuildOptions, client docker.Docker) error {
	buildArgs := buildArgs(opts)
	pushArgs := pushArgs(opts)
	fmt.Println("Building image ...")
	if err := client.Exec("build", buildArgs[1:]...); err != nil {
		return err
	}
	if opts.Push {
		fmt.Println("Pushing image...")
		return client.Exec("push", pushArgs[1:]...)
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

func validateBuildOptions(options *BuildOptions) error {
	if err := validateFilepath(&options.FilePath); err != nil {
		return err
	}
	err := validateFunctionName(&options.FunctionName, options.FilePath)
	return err
}
