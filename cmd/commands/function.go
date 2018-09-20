/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"fmt"

	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

const (
	functionCreateInvokerIndex = iota
	functionCreateFunctionNameIndex
	functionCreateNumberOfArgs
)

const (
	functionBuildFunctionNameIndex = iota
	functionBuildNumberOfArgs
)

func Function() *cobra.Command {
	return &cobra.Command{
		Use:   "function",
		Short: "Interact with function related resources",
	}
}

func FunctionCreate(fcTool *core.Client) *cobra.Command {
	createFunctionOptions := core.CreateFunctionOptions{}

	// runtime definitions
	buildpacks := map[string]string{
		"java": "projectriff/buildpack",
	}
	invokers := map[string]string{
		"jar":     "https://github.com/projectriff/java-function-invoker/raw/v0.1.1/java-invoker.yaml",
		"command": "https://github.com/projectriff/command-function-invoker/raw/v0.0.7/command-invoker.yaml",
		"node":    "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
	}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a new function resource",
		Long: "Create a new function resource from the content of the provided Git repo/revision or local source.\n" +
			"\nThe INVOKER arg defines the language runtime and function invoker that is added to the function code in the build step. The resulting image is then used to create a Knative Service (`service.serving.knative.dev`) instance of the name specified for the function. The following invokers are available:\n\n" +
			"- 'java': uses the riff Buildpack to build Maven or Gradle projects from source\n" +
			"- 'jar': uses riff's java-function-invoker build for a prebuilt JAR file\n" +
			"- 'node': uses riff's node-function-invoker build\n" +
			"- 'command': uses riff's command-function-invoker build\n" +
			"\nBuildpack based builds support building from local source or within the cluster. Images will be pushed to the registry specified in the image name, unless prefixed with 'dev.local/' in which case the image will only be available within the local Docker daemon.\n" +
			"\nFrom then on you can use the sub-commands for the `service` command to interact with the service created for the function.\n\n" +
			envFromLongDesc + "\n",
		Example: `  riff function create node square --git-repo https://github.com/acme/square --image acme/square --namespace joseph-ns
  riff function create java tweets-logger --git-repo https://github.com/acme/tweets --image acme/tweets-logger:1.0.0`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionCreateNumberOfArgs),
			AtPosition(functionCreateInvokerIndex, ValidName()),
			AtPosition(functionCreateFunctionNameIndex, ValidName()),
		),
		PreRunE: FlagsValidatorAsCobraRunE(AtLeastOneOf("git-repo", "local-path")),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[functionCreateFunctionNameIndex]

			invoker := args[functionCreateInvokerIndex]
			createFunctionOptions.Invoker = invoker
			if buildpack, exists := buildpacks[invoker]; exists {
				createFunctionOptions.BuildpackImage = buildpack
			} else if invokerURL, exists := invokers[invoker]; exists {
				createFunctionOptions.InvokerURL = invokerURL
			} else {
				return fmt.Errorf("unknown invoker: %s", invoker)
			}

			createFunctionOptions.Name = fnName
			f, err := (*fcTool).CreateFunction(createFunctionOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			if createFunctionOptions.DryRun {
				marshaller := NewMarshaller(cmd.OutOrStdout())
				if err = marshaller.Marshal(f); err != nil {
					return err
				}
			} else {
				PrintSuccessfulCompletion(cmd)
				if !createFunctionOptions.Verbose && !createFunctionOptions.Wait {
					namespaceOption := ""
					if createFunctionOptions.Namespace != "" {
						namespaceOption = fmt.Sprintf(" -n %s", createFunctionOptions.Namespace)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Issue `riff service status %s%s` to see the status of the function\n", fnName, namespaceOption)
				}
			}

			return nil
		},
	}

	LabelArgs(command, "INVOKER", "FUNCTION_NAME")

	command.Flags().StringVarP(&createFunctionOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")
	command.Flags().BoolVarP(&createFunctionOptions.DryRun, "dry-run", "", false, dryRunUsage)
	command.Flags().StringVar(&createFunctionOptions.Image, "image", "", "the name of the image to build; must be a writable `repository/image[:tag]` with credentials configured")
	command.MarkFlagRequired("image")
	command.Flags().StringVar(&createFunctionOptions.GitRepo, "git-repo", "", "the `URL` for a git repository hosting the function code")
	command.Flags().StringVar(&createFunctionOptions.GitRevision, "git-revision", "master", "the git `ref-spec` of the function code to use")
	command.Flags().StringVarP(&createFunctionOptions.LocalPath, "local-path", "l", "", "`path` to local source to build the image from")
	command.Flags().StringVar(&createFunctionOptions.Handler, "handler", "", "the name of the `method or class` to invoke, depending on the invoker used")
	command.Flags().StringVar(&createFunctionOptions.Artifact, "artifact", "", "`path` to the function source code or jar file; auto-detected if not specified")
	command.Flags().BoolVarP(&createFunctionOptions.Verbose, "verbose", "v", false, verboseUsage)
	command.Flags().BoolVarP(&createFunctionOptions.Wait, "wait", "w", false, waitUsage)

	command.Flags().StringArrayVar(&createFunctionOptions.Env, "env", []string{}, envUsage)
	command.Flags().StringArrayVar(&createFunctionOptions.EnvFrom, "env-from", []string{}, envFromUsage)

	return command
}

func FunctionBuild(fcTool *core.Client) *cobra.Command {

	buildFunctionOptions := core.BuildFunctionOptions{}

	command := &cobra.Command{
		Use:     "build",
		Short:   "Trigger a revision build for a function resource",
		Example: `  riff function build square`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionBuildNumberOfArgs),
			AtPosition(functionBuildFunctionNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			fnName := args[functionBuildFunctionNameIndex]

			buildFunctionOptions.Name = fnName
			err := (*fcTool).BuildFunction(buildFunctionOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)

			return nil
		},
	}

	LabelArgs(command, "FUNCTION_NAME")

	command.Flags().StringVarP(&buildFunctionOptions.Namespace, "namespace", "n", "", "the `namespace` of the function")
	command.Flags().StringVarP(&buildFunctionOptions.LocalPath, "local-path", "l", "", "path to local source to build the image from")
	command.Flags().BoolVarP(&buildFunctionOptions.Verbose, "verbose", "v", false, verboseUsage)
	command.Flags().BoolVarP(&buildFunctionOptions.Wait, "wait", "w", false, waitUsage)

	return command
}
