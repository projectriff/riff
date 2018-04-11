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

	"path/filepath"

	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/pkg/initializer"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Init(invokers []projectriff_v1.Invoker) (*cobra.Command, *options.InitOptions) {

	var initOptions = options.InitOptions{}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a function",
		Long:  utils.InitCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			if len(invokers) == 0 {
				return fmt.Errorf("Invokers must be installed, run `riff invokers apply --help` for help")
			}
			names := invokerNames(invokers)
			return fmt.Errorf("The invoker must be specified. Pick one of: %s", strings.Join(names, ", "))
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			initOptions.UserAccount = utils.GetUseraccountWithOverride("useraccount", *cmd.Flags())
			err := validateInitOptions(&initOptions)
			if err != nil {
				return err
			}
			return nil

		},
	}

	initCmd.PersistentFlags().BoolVar(&initOptions.DryRun, "dry-run", false, "print generated function artifacts content to stdout only")
	initCmd.PersistentFlags().StringVarP(&initOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	initCmd.PersistentFlags().StringVarP(&initOptions.FunctionName, "name", "n", "", "the name of the function (defaults to the name of the current directory)")
	initCmd.PersistentFlags().StringVarP(&initOptions.Version, "version", "v", utils.DefaultValues.Version, "the version of the function image")
	initCmd.PersistentFlags().StringVar(&initOptions.InvokerVersion, "invoker-version", "", "the version of the invoker to use when building containers")
	initCmd.PersistentFlags().StringVarP(&initOptions.UserAccount, "useraccount", "u", utils.DefaultValues.UserAccount, "the Docker user account to be used for the image repository")
	initCmd.PersistentFlags().StringVarP(&initOptions.Artifact, "artifact", "a", "", "path to the function artifact, source code or jar file")
	initCmd.PersistentFlags().StringVarP(&initOptions.Input, "input", "i", "", "the name of the input topic (defaults to function name)")
	initCmd.PersistentFlags().StringVarP(&initOptions.Output, "output", "o", "", "the name of the output topic (optional)")
	initCmd.PersistentFlags().BoolVar(&initOptions.Force, "force", utils.DefaultValues.Force, "overwrite existing functions artifacts")

	initCmd.SetUsageTemplate(utils.CustomInvokerUsageTemplate)

	return initCmd, &initOptions
}

func InitInvokers(invokers []projectriff_v1.Invoker, initOptions *options.InitOptions) ([]*cobra.Command, error) {

	var initInvokerCmds []*cobra.Command
	for _, invoker := range invokers {
		invokerName := invoker.ObjectMeta.Name
		var initInvokerCmd = &cobra.Command{
			Use:   invokerName,
			Short: fmt.Sprintf("Initialize a %s function", invokerName),
			Long:  utils.InitInvokerCmdLong(invoker),
			Args:  utils.AliasFlagToSoleArg("filepath"),
			RunE: func(cmd *cobra.Command, args []string) error {
				invoker, err := invokerForName(invokerName, invokers)
				if err != nil {
					return err
				}
				return initializer.Initialize(invoker, initOptions)
			},
		}

		handler := invoker.Spec.Handler
		if handler.Default != "" || handler.Description != "" {
			initInvokerCmd.Flags().StringVar(&initOptions.Handler, "handler", handler.Default, handler.Description)
			if handler.Default == "" {
				initInvokerCmd.MarkFlagRequired("handler")
			}
		}

		initInvokerCmds = append(initInvokerCmds, initInvokerCmd)
	}
	return initInvokerCmds, nil
}

func validateInitOptions(options *options.InitOptions) error {
	options.FilePath = filepath.Clean(options.FilePath)
	if err := validateFunctionName(&options.FunctionName, options.FilePath); err != nil {
		return err
	}

	if err := validateAndCleanArtifact(&options.Artifact, options.FilePath); err != nil {
		return err
	}

	if err := validateProtocol(&options.Protocol); err != nil {
		return err
	}
	return nil
}

func invokerForName(name string, invokers []projectriff_v1.Invoker) (projectriff_v1.Invoker, error) {
	for _, invoker := range invokers {
		if invoker.ObjectMeta.Name == name {
			return invoker, nil
		}
	}
	return projectriff_v1.Invoker{}, fmt.Errorf("No invoker found for %s", name)
}

func invokerNames(invokers []projectriff_v1.Invoker) []string {
	names := []string{}
	for _, invoker := range invokers {
		names = append(names, invoker.ObjectMeta.Name)
	}
	return names
}
