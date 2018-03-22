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
	"strings"

	"github.com/projectriff/riff/riff-cli/pkg/functions"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/projectriff/riff/riff-cli/pkg/jsonpath"
	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
)

type DeleteOptions struct {
	FilePath     string
	FunctionName string
	Namespace    string
	DryRun       bool
	All          bool
}

func Delete() (*cobra.Command, *DeleteOptions) {

	deleteOptions := DeleteOptions{}

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete function resources in the cluster",
		Long:  `Delete the resource[s] for the function or path specified.`,
		Example: `  riff delete -n square
    or
  riff delete -f function/square`,
		Args: utils.AliasFlagToSoleArg("filepath"),

		RunE: func(cmd *cobra.Command, args []string) error {
			return doDelete(cmd, deleteOptions)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {

			deleteOptions.Namespace = utils.GetStringValueWithOverride("namespace", *cmd.Flags())

			// If name and no file path given, skip this step
			if deleteOptions.FilePath != "" && deleteOptions.FunctionName == "" {
				return  options.ValidateNamePathOptions(&deleteOptions.FunctionName, &deleteOptions.FilePath)
			}
			return nil
		},
	}

	deleteCmd.Flags().BoolVar(&deleteOptions.All, "all", false, "delete all resources including topics, not just the function resource")
	deleteCmd.Flags().BoolVar(&deleteOptions.DryRun, "dry-run", false, "print generated function artifacts content to stdout only")
	deleteCmd.Flags().StringVarP(&deleteOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	deleteCmd.Flags().StringVarP(&deleteOptions.FunctionName, "name", "n", "", "the name of the function (defaults to the name of the current directory)")
	deleteCmd.Flags().StringVar(&deleteOptions.Namespace, "namespace", "", "the namespace used for the deployed resources (defaults to kubectl's default)")

	return deleteCmd, &deleteOptions
}

func doDelete(cmd *cobra.Command, opts DeleteOptions) error {

	var cmdArgs []string
	var message string

	if opts.FilePath == "" && opts.FunctionName != "" {
		err := deleteFunctionByName(opts)
		if err != nil {
			cmd.SilenceUsage = true
		}
		return err
	}

	var err error
	opts.FunctionName, err = functions.FunctionNameFromPath(opts.FilePath)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	abs, err := functions.AbsPath(opts.FilePath)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	cmdArgs = []string{"delete"}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}

	if opts.All {
		optionPath := opts.FilePath
		if !osutils.IsDirectory(abs) {
			abs = filepath.Dir(abs)
			optionPath = filepath.Dir(optionPath)
		}
		message = fmt.Sprintf("Deleting resources %v\n\n", optionPath)
		resourceDefinitionPaths, err := osutils.FindRiffResourceDefinitionPaths(abs)
		if err != nil {
			return err
		}

		if len(resourceDefinitionPaths) > 0 {
			for _, resourceDefinitionPath := range resourceDefinitionPaths {
				cmdArgs = append(cmdArgs, "-f", resourceDefinitionPath)
			}
		} else {
			fmt.Printf("No resources found for path %s\n", abs)
			return nil
		}
	} else {
		if osutils.IsDirectory(abs) {
			message = fmt.Sprintf("Deleting function %v\n\n", opts.FunctionName)
			cmdArgs = append(cmdArgs, "function", opts.FunctionName)
		} else {
			message = fmt.Sprintf("Deleting resource %v\n\n", opts.FilePath)
			cmdArgs = append(cmdArgs, "-f", opts.FilePath)
		}
	}

	err = deleteResources(cmdArgs, message, opts.DryRun)

	if err != nil {
		cmd.SilenceUsage = true
	}
	return err
}

func deleteFunctionByName(opts DeleteOptions) error {
	getArgs := []string{"get"}
	if opts.Namespace != "" {
		getArgs = append(getArgs, "--namespace", opts.Namespace)
	}
	getArgs = append(getArgs, "function", opts.FunctionName, "-o", "json")
	json, err := kubectl.EXEC_FOR_BYTES(getArgs)
	if err != nil {
		return err
	}

	err = deleteFunction(opts.FunctionName, opts)
	if err != nil {
		return err
	}

	if opts.All {
		parser := jsonpath.NewParser(json)
		inputTopic := parser.Value(`$.spec.input+`)
		outputTopic := parser.Value(`$.spec.output+`)
		if inputTopic != "" {
			err = deleteTopic(inputTopic, opts)
		}
		if outputTopic != "" {
			err = deleteTopic(outputTopic, opts)
		}
	}
	return err
}

func deleteTopic(topic string, opts DeleteOptions) error {
	cmdArgs := []string{"delete"}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}
	cmdArgs = append(cmdArgs, "topic", topic)
	return deleteResources(cmdArgs, fmt.Sprintf("Deleting topic %v\n\n", topic), opts.DryRun)
}

func deleteFunction(function string, opts DeleteOptions) error {
	cmdArgs := []string{"delete"}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}
	cmdArgs = append(cmdArgs, "function", function)
	return deleteResources(cmdArgs, fmt.Sprintf("Deleting function %v\n\n", function), opts.DryRun)
}

func deleteResources(cmdArgs []string, message string, dryRun bool) error {
	if dryRun {
		fmt.Printf("\nDelete Command: kubectl %s\n\n", strings.Trim(fmt.Sprint(cmdArgs), "[]"))
	} else {
		fmt.Print(message)
		output, err := kubectl.EXEC_FOR_STRING(cmdArgs)
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", output)
	}
	return nil
}
