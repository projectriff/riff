/*
 * Copyright 2018-2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
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
	"strings"

	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
)

const (
	functionCreateFunctionNameIndex = iota
	functionCreateNumberOfArgs
)

const (
	functionUpdateFunctionNameIndex = iota
	functionUpdateNumberOfArgs
)

const (
	functionBuildNumberOfArgs = iota
)

func Function() *cobra.Command {
	return &cobra.Command{
		Use:   "function",
		Short: "Interact with function related resources",
	}
}

func FunctionCreate(buildpackBuilder core.Builder, fcTool *core.Client) *cobra.Command {
	createFunctionOptions := core.CreateFunctionOptions{}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a new function resource",
		Long: "Create a new function resource from the content of the provided Git repo/revision or local source.\n\n" +
			"The --invoker flag can be used to force the language runtime and function invoker that is added to the function code in the build step. The resulting image is then used to create a Knative Service (`service.serving.knative.dev`) instance of the name specified for the function.\n\n" +
			"Images will be pushed to the registry specified in the image name. If a default image prefix was specified during namespace init, the image flag is optional. The function name is combined with the default prefix to define the image. Instead of using the function name, a custom repository can be specified with the image set like `--image _/custom-name` which would resolve to `docker.io/example/custom-name`.\n\n" +
			"From then on you can use the sub-commands for the `service` command to interact with the service created for the function.\n\n" +
			envFromLongDesc + "\n",
		Example: `  ` + env.Cli.Name + ` function create square --git-repo https://github.com/acme/square --artifact square.js --image acme/square --invoker node --namespace joseph-ns
  ` + env.Cli.Name + ` function create tweets-logger --git-repo https://github.com/acme/tweets --image acme/tweets-logger:1.0.0`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			validator := FlagsValidatorAsCobraRunE(AtLeastOneOf("git-repo", "local-path"))
			err := validator(cmd, args)
			if err != nil {
				return err
			}

			if createFunctionOptions.Image == "" || strings.HasPrefix(createFunctionOptions.Image, "_") {
				prefix, err := (*fcTool).DefaultBuildImagePrefix(createFunctionOptions.Namespace)
				if err != nil {
					return fmt.Errorf("unable to default image: %s", err)
				}
				if prefix == "" {
					if createFunctionOptions.Image == "" {
						return fmt.Errorf("required flag(s) \"image\" not set, this flag is optional if --image-prefix is specified during namespace init")
					}
					return fmt.Errorf("--image flag must include a repository, the image prefix was not set during namespace init")
				}

				fnName := args[functionCreateFunctionNameIndex]
				if createFunctionOptions.Image == "" {
					// combine prefix and function name to provide default image
					createFunctionOptions.Image = fmt.Sprintf("%s/%s", prefix, fnName)
				} else if strings.HasPrefix(createFunctionOptions.Image, "_/") {
					// add the prefix to the specified image name
					createFunctionOptions.Image = strings.Replace(createFunctionOptions.Image, "_", prefix, 1)
				} else {
					return fmt.Errorf("Unknown image prefix syntax, expected %q, found: %q", "_/", createFunctionOptions.Image)
				}
			}

			return nil
		},
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionCreateNumberOfArgs),
			AtPosition(functionCreateFunctionNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[functionCreateFunctionNameIndex]

			createFunctionOptions.Name = fnName
			f, pvc, err := (*fcTool).CreateFunction(buildpackBuilder, createFunctionOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			if createFunctionOptions.DryRun {
				marshaller := NewMarshaller(cmd.OutOrStdout())
				if err = marshaller.Marshal(pvc); err != nil {
					return err
				}
				if err = marshaller.Marshal(f); err != nil {
					return err
				}
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n%s function create: built image %s for service %s.\n", env.Cli.Name, f.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image, f.Name)
				PrintSuccessfulCompletion(cmd)

				if !createFunctionOptions.Verbose && !createFunctionOptions.Wait {
					namespaceOption := ""
					if createFunctionOptions.Namespace != "" {
						namespaceOption = fmt.Sprintf(" -n %s", createFunctionOptions.Namespace)
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Issue `%s service status %s%s` to see the status of the function\n", env.Cli.Name, fnName, namespaceOption)
				}
			}

			return nil
		},
	}

	LabelArgs(command, "FUNCTION_NAME")

	registerBuildOptionsFlags(command, &createFunctionOptions.BuildOptions)
	command.Flags().StringVarP(&createFunctionOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")
	command.Flags().BoolVarP(&createFunctionOptions.DryRun, "dry-run", "", false, dryRunUsage)
	command.Flags().StringVar(&createFunctionOptions.Image, "image", "", "the name of the image to build; must be a writable `repository/image[:tag]` with credentials configured")
	command.Flags().StringVar(&createFunctionOptions.GitRepo, "git-repo", "", "the `URL` for a git repository hosting the function code")
	command.Flags().StringVar(&createFunctionOptions.GitRevision, "git-revision", "master", "the git `ref-spec` of the function code to use")
	command.Flags().StringVar(&createFunctionOptions.SubPath, "sub-path", "", "the directory within the git repo to expose, files outside of this directory will not be available during the build")
	command.Flags().StringVarP(&createFunctionOptions.LocalPath, "local-path", "l", "", "`path` to local source to build the image from; only build-pack builds are supported at this time")
	command.Flags().BoolVarP(&createFunctionOptions.Verbose, "verbose", "v", false, verboseUsage)
	command.Flags().BoolVarP(&createFunctionOptions.Wait, "wait", "w", false, waitUsage)

	command.Flags().StringArrayVar(&createFunctionOptions.Env, "env", []string{}, envUsage)
	command.Flags().StringArrayVar(&createFunctionOptions.EnvFrom, "env-from", []string{}, envFromUsage)

	return command
}

func FunctionUpdate(buildpackBuilder core.Builder, fcTool *core.Client) *cobra.Command {

	updateFunctionOptions := core.UpdateFunctionOptions{}

	command := &cobra.Command{
		Use:     "update",
		Short:   "Trigger a build to generate a new revision for a function resource",
		Example: `  ` + env.Cli.Name + ` function update square`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionUpdateNumberOfArgs),
			AtPosition(functionUpdateFunctionNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			fnName := args[functionUpdateFunctionNameIndex]

			updateFunctionOptions.Name = fnName
			err := (*fcTool).UpdateFunction(buildpackBuilder, updateFunctionOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)

			return nil
		},
	}

	LabelArgs(command, "FUNCTION_NAME")

	command.Flags().StringVarP(&updateFunctionOptions.Namespace, "namespace", "n", "", "the `namespace` of the function")
	command.Flags().StringVarP(&updateFunctionOptions.LocalPath, "local-path", "l", "", "path to local source to build the image from; only build-pack builds are supported at this time")
	command.Flags().BoolVarP(&updateFunctionOptions.Verbose, "verbose", "v", false, verboseUsage)
	command.Flags().BoolVarP(&updateFunctionOptions.Wait, "wait", "w", false, waitUsage)

	return command
}

func FunctionBuild(buildpackBuilder core.Builder, fcTool *core.Client) *cobra.Command {
	buildFunctionOptions := core.BuildFunctionOptions{}

	command := &cobra.Command{
		Use:   "build",
		Short: "Build a function container from local source",
		Args:  cobra.ExactArgs(functionBuildNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*fcTool).BuildFunction(buildpackBuilder, buildFunctionOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)

			return nil
		},
	}

	registerBuildOptionsFlags(command, &buildFunctionOptions.BuildOptions)
	command.Flags().StringVar(&buildFunctionOptions.Image, "image", "", "the name of the image to build; must be a writable `repository/image[:tag]` with credentials configured")
	command.MarkFlagRequired("image")
	command.Flags().StringVarP(&buildFunctionOptions.LocalPath, "local-path", "l", "", "`path` to local source to build the image from; only build-pack builds are supported at this time")
	command.MarkFlagRequired("local-path")

	return command
}

func registerBuildOptionsFlags(command *cobra.Command, options *core.BuildOptions) {
	command.Flags().StringVar(&options.Invoker, "invoker", "", "invoker runtime to override `language` detected by buildpack")
	command.Flags().StringVar(&options.Handler, "handler", "", "the name of the `method or class` to invoke, depending on the invoker used")
	command.Flags().StringVar(&options.Artifact, "artifact", "", "`path` to the function source code or jar file; auto-detected if not specified")
}
