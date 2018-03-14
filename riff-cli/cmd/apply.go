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
	"strings"

	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/functions"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
	"errors"
)

type ApplyOptions struct {
	FilePath  string
	Namespace string
	DryRun    bool
}

func Apply() (*cobra.Command, *ApplyOptions) {

	applyOptions := ApplyOptions{}

	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply function resource definitions",
		Long:  `Apply the resource definition[s] included in the path. A resource will be created if it doesn't exist yet.`,
		Example: `  riff apply -f some/function/path
  riff apply -f some/function/path/some.yaml`,

		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return apply(cmd, applyOptions)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			applyOptions.Namespace = utils.GetStringValueWithOverride("namespace", *cmd.Flags())

			//TODO: DRY
			if len(args) > 0 {
				if len(args) == 1 && applyOptions.FilePath == "" {
					applyOptions.FilePath = args[0]
				} else {
					return errors.New(fmt.Sprintf("Invalid argument(s) %v\n", args))
				}
			}
			return validateApplyOptions(&applyOptions)
		},
	}

	applyCmd.Flags().BoolVar(&applyOptions.DryRun, "dry-run", false, "print generated function artifacts content to stdout only")
	applyCmd.Flags().StringVarP(&applyOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	applyCmd.Flags().StringVar(&applyOptions.Namespace, "namespace", "", "the namespace used for the deployed resources (defaults to kubectl's default)")

	return applyCmd, &applyOptions
}

func apply(cmd *cobra.Command, opts ApplyOptions) error {
	abs, err := functions.AbsPath(opts.FilePath)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	cmdArgs := []string{"apply"}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}

	var message string

	if osutils.IsDirectory(abs) {
		message = fmt.Sprintf("Applying resources in %v\n\n", opts.FilePath)
		resourceDefinitionPaths, err := osutils.FindRiffResourceDefinitionPaths(abs)
		if err != nil {
			return err
		}
		for _, resourceDefinitionPath := range resourceDefinitionPaths {
			cmdArgs = append(cmdArgs, "-f", resourceDefinitionPath)
		}
	} else {
		message = fmt.Sprintf("Applying resource %v\n\n", opts.FilePath)
		cmdArgs = append(cmdArgs, "-f", abs)
	}

	if opts.DryRun {
		fmt.Printf("\nApply Command: kubectl %s\n\n", strings.Trim(fmt.Sprint(cmdArgs), "[]"))
	} else {
		fmt.Print(message)
		output, err := kubectl.ExecForString(cmdArgs)
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}
		fmt.Printf("%v\n", output)
	}
	return nil
}

func validateApplyOptions(options *ApplyOptions) error {
	return validateFilepath(&options.FilePath)
}
