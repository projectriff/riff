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
	"path/filepath"
	"strings"

	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/functions"
	"github.com/projectriff/riff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/projectriff/riff/riff-cli/pkg/jsonpath"
	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/cmd/opts"
)


func Delete() *cobra.Command {

	// deleteCmd represents the delete command
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete function resources",
		Long:  `Delete the resource[s] for the function or path specified.`,
		Example: `  riff delete -n square
    or
  riff delete -f function/square`,

		RunE: func(cmd *cobra.Command, args []string) error {

			return delete(cmd, options.GetDeleteOptions(opts.DeleteAllOptions))

		},
		PreRun: func(cmd *cobra.Command, args []string) {

			if !opts.DeleteAllOptions.Initialized {
				utils.MergeDeleteOptions(*cmd.Flags(), &opts.DeleteAllOptions)

				if len(args) > 0 {
					if len(args) == 1 && opts.DeleteAllOptions.FilePath == "" {
						opts.DeleteAllOptions.FilePath = args[0]
					} else {
						ioutils.Errorf("Invalid argument(s) %v\n", args)
						cmd.Usage()
						os.Exit(1)
					}
				}
				/*
				 * If name and no file path given, skip this step
				 */
				if opts.DeleteAllOptions.FilePath != "" && opts.DeleteAllOptions.FunctionName == "" {
					err := options.ValidateNamePathOptions(&opts.DeleteAllOptions.FunctionName, &opts.DeleteAllOptions.FilePath)
					if err != nil {
						ioutils.Error(err)
						os.Exit(1)
					}
				}
			}
			opts.DeleteAllOptions.Initialized = true
		},
	}
	utils.CreateDeleteFlags(deleteCmd.Flags())
	return deleteCmd
}

func delete(cmd *cobra.Command, opts options.DeleteOptions) error {

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

func deleteFunctionByName(opts options.DeleteOptions) error {
	json, err := kubectl.EXEC_FOR_BYTES([]string{"get", "function", opts.FunctionName, "-o", "json"})
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

func deleteTopic(topic string, opts options.DeleteOptions) error {
	cmdArgs := []string{"delete"}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}
	cmdArgs = append(cmdArgs, "topic", topic)
	return deleteResources(cmdArgs, fmt.Sprintf("Deleting topic %v\n\n", topic), opts.DryRun)
}

func deleteFunction(function string, opts options.DeleteOptions) error {
	cmdArgs := []string{"delete"}
	if opts.Namespace != "" {
		cmdArgs = append(cmdArgs, "--namespace", opts.Namespace)
	}
	cmdArgs = append(cmdArgs, "function", function)
	return deleteResources(cmdArgs, fmt.Sprintf("Deleting function %v\n\n", function), opts.DryRun)
}

func deleteResources(cmdArgs []string, message string, dryRun bool) error {
	if (dryRun) {
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
