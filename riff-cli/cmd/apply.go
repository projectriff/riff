/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        https://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
)

type ApplyOptions struct {
	FilePath  string
	Namespace string
	DryRun    bool
}

func Apply(realKubeCtl kubectl.KubeCtl, dryRunKubeCtl kubectl.KubeCtl) (*cobra.Command, *ApplyOptions) {

	applyOptions := ApplyOptions{}

	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply function resource definitions to the cluster",
		Long:  `Apply the resource definition[s] included in the path. A resource will be created if it doesn't exist yet.`,
		Example: `  riff apply -f some/function/path
  riff apply -f some/function/path/some.yaml`,
		Args: utils.AliasFlagToSoleArg("filepath"),

		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			kubectlClient := realKubeCtl
			if applyOptions.DryRun {
				kubectlClient = dryRunKubeCtl
			}
			return apply(applyOptions, kubectlClient)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			applyOptions.Namespace = utils.GetStringValueWithOverride("namespace", *cmd.Flags())
			applyOptions.FilePath = filepath.Clean(applyOptions.FilePath)
			return nil
		},
	}

	applyCmd.Flags().BoolVar(&applyOptions.DryRun, "dry-run", false, "print generated function artifacts content to stdout only")
	applyCmd.Flags().StringVarP(&applyOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	applyCmd.Flags().StringVar(&applyOptions.Namespace, "namespace", "", "the namespace used for the deployed resources (defaults to kubectl's default)")

	return applyCmd, &applyOptions
}

func apply(opts ApplyOptions, kubectlClient kubectl.KubeCtl) error {
	abs, err := osutils.AbsPath(opts.FilePath)
	if err != nil {
		return err
	}

	cmdArgs := []string{"apply"}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}

	var message string
	var resourceDefinitionPaths []string

	if osutils.IsDirectory(abs) {
		message = fmt.Sprintf("Applying resources in %v\n\n", opts.FilePath)
		resourceDefinitionPaths, err = osutils.FindRiffResourceDefinitionPaths(abs)
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
	if len(resourceDefinitionPaths) == 0 {
		fmt.Printf("No riff resources found in %v\n", opts.FilePath)
		return nil
	}
	fmt.Print(message)
	output, err := kubectlClient.Exec(cmdArgs)
	fmt.Printf("%v\n", output)
	return err
}
