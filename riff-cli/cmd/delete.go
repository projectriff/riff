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

	"os"

	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/functions"
	"github.com/projectriff/riff/riff-cli/pkg/jsonpath"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	FunctionName string
	Namespace    string
	DryRun       bool
	All          bool
}

func Delete(realKubeCtl kubectl.KubeCtl, dryRunKubeCtl kubectl.KubeCtl) (*cobra.Command, *DeleteOptions) {

	deleteOptions := DeleteOptions{}

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete function resources in the cluster",
		Long:  `Delete the resource[s] for the function or path specified.`,
		Example: `  riff delete -n square
    or
  riff delete -f function/square`,

		RunE: func(cmd *cobra.Command, args []string) error {
			// From this point on, errors are execution errors, not misuse of flags
			cmd.SilenceUsage = true
			actingClient := realKubeCtl
			if deleteOptions.DryRun {
				actingClient = dryRunKubeCtl
			}
			return deleteFunctionByName(deleteOptions, actingClient, realKubeCtl)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {

			deleteOptions.Namespace = utils.GetStringValueWithOverride("namespace", *cmd.Flags())
			if deleteOptions.FunctionName == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				if deleteOptions.FunctionName, err = functions.FunctionNameFromPath(cwd); err != nil {
					return err
				}
			}
			return nil
		},
	}

	deleteCmd.Flags().BoolVar(&deleteOptions.All, "all", false, "delete all resources including topics, not just the function resource")
	deleteCmd.Flags().BoolVar(&deleteOptions.DryRun, "dry-run", false, "print generated commands to stdout only")
	deleteCmd.Flags().StringVarP(&deleteOptions.FunctionName, "name", "n", "", "the name of the function (defaults to the name of the current directory)")
	deleteCmd.Flags().StringVar(&deleteOptions.Namespace, "namespace", "", "the namespace used for the deployed resources (defaults to kubectl's default)")

	return deleteCmd, &deleteOptions
}

func deleteFunctionByName(opts DeleteOptions, actingClient kubectl.KubeCtl, queryClient kubectl.KubeCtl) error {
	if opts.All {

		if inputTopic, outputTopic, err := lookupTopicNames(opts, queryClient); err != nil {
			return err
		} else {
			cmdArgs := []string{"delete", "topics.projectriff.io", "<placeholder>"}
			if opts.Namespace != "" {
				cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
			}
			if inputTopic != "" {
				cmdArgs[2] = inputTopic
				deleteResources(cmdArgs, actingClient)
			}
			if outputTopic != "" {
				cmdArgs[2] = outputTopic
				deleteResources(cmdArgs, actingClient)
			}
		}
	}

	cmdArgs := []string{"delete", "functions.projectriff.io", opts.FunctionName}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}
	if err := deleteResources(cmdArgs, actingClient); err != nil {
		return err
	}

	return nil
}

func deleteResources(cmdArgs []string, actingClient kubectl.KubeCtl) error {
	fmt.Printf("Deleting %v %v:\n", cmdArgs[1], cmdArgs[2])
	output, err := actingClient.Exec(cmdArgs)
	fmt.Println(output)
	return err
}

func lookupTopicNames(opts DeleteOptions, queryClient kubectl.KubeCtl) (string, string, error) {
	getArgs := []string{"get"}
	if opts.Namespace != "" {
		getArgs = append(getArgs, "--namespace", opts.Namespace)
	}
	getArgs = append(getArgs, "functions.projectriff.io", opts.FunctionName, "-o", "json")
	output, err := queryClient.Exec(getArgs)
	if err != nil {
		fmt.Println(output)
		return "", "", err
	}
	parser := jsonpath.NewParser([]byte(output))

	inputTopic, _ := parser.StringValue(`$.spec.input`)
	outputTopic, _ := parser.StringValue(`$.spec.output`)
	return inputTopic, outputTopic, err

}
