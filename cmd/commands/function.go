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
	functionCreateRuntimeIndex = iota
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
		"java-buildpack": "projectriff/buildpack",
		"detect":         "projectriff/buildpack",
	}
	invokers := map[string]string{
		"java":            "https://github.com/projectriff/java-function-invoker/raw/v0.0.7/java-invoker.yaml",
		"java-invoker":    "https://github.com/projectriff/java-function-invoker/raw/v0.0.7/java-invoker.yaml",
		"command":         "https://github.com/projectriff/command-function-invoker/raw/v0.0.7/command-invoker.yaml",
		"command-invoker": "https://github.com/projectriff/command-function-invoker/raw/v0.0.7/command-invoker.yaml",
		"node":            "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
		"node-invoker":    "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
	}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a new function resource",
		Long: "Create a new function resource from the content of the provided Git repo/revision or local source.\n" +
			"\nThe RUNTIME arg defines the language runtime that is added to the function code in the build step. The resulting image is then used to create a Knative Service (`service.serving.knative.dev`) instance of the name specified for the function. The following runtimes are available:\n\n" +
			// TODO make runtime help data-driven
			"- 'java': uses riff's java-function-invoker (aliased as java-invoker)\n" +
			"- 'node': uses riff's node-function-invoker (aliased as node-invoker)\n" +
			"- 'command': uses riff's command-function-invoker (aliased as command-invoker)\n" +
			"- 'java-buildpack': uses the riff Buildpack \n" +
			"- 'detect': uses the riff Buildpack's detection (currently limited to Java functions) \n" +
			"\nClassic riff Invoker runtimes are available in addition to experimental Buildpack runtimes.\n" +
			"\nBuildpack based runtimes support building from local source in addition to within the cluster. Locally built images prefixed with 'dev.local/' are saved to the local Docker daemon while all other images are pushed to the registry specified in the image name.\n" +
			"\nFrom then on you can use the sub-commands for the `service` command to interact with the service created for the function.\n\n" +
			envFromLongDesc + "\n",
		Example: `  riff function create node square --git-repo https://github.com/acme/square --image acme/square --namespace joseph-ns
  riff function create java tweets-logger --git-repo https://github.com/acme/tweets --image acme/tweets-logger:1.0.0`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionCreateNumberOfArgs),
			AtPosition(functionCreateRuntimeIndex, ValidName()),
			AtPosition(functionCreateFunctionNameIndex, ValidName()),
		),
		PreRunE: FlagsValidatorAsCobraRunE(AtLeastOneOf("git-repo", "local-path")),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[functionCreateFunctionNameIndex]

			runtime := args[functionCreateRuntimeIndex]
			if buildpack, exists := buildpacks[runtime]; exists {
				createFunctionOptions.BuildpackImage = buildpack
			} else if invokerURL, exists := invokers[runtime]; exists {
				createFunctionOptions.InvokerURL = invokerURL
			} else {
				return fmt.Errorf("unknown runtime: %s", runtime)
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

	LabelArgs(command, "RUNTIME", "FUNCTION_NAME")

	command.Flags().VarP(
		BroadcastStringValue("",
			&createFunctionOptions.Namespace,
		),
		"namespace", "n", "the `namespace` of the subscription, channel, and function",
	)

	command.Flags().VarPF(
		BroadcastBoolValue(false,
			&createFunctionOptions.DryRun,
		),
		"dry-run", "", dryRunUsage,
	).NoOptDefVal = "true"

	command.Flags().StringVar(&createFunctionOptions.Image, "image", "", "the name of the image to build; must be a writable `repository/image[:tag]` with credentials configured")
	command.MarkFlagRequired("image")
	command.Flags().StringVar(&createFunctionOptions.GitRepo, "git-repo", "", "the `URL` for a git repository hosting the function code")
	command.Flags().StringVar(&createFunctionOptions.GitRevision, "git-revision", "master", "the git `ref-spec` of the function code to use")
	command.Flags().StringVarP(&createFunctionOptions.LocalPath, "local-path", "l", "", "path to local source to build the image from")
	command.Flags().StringVar(&createFunctionOptions.Handler, "handler", "", "the name of the `method or class` to invoke, depending on the runtime used")
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
