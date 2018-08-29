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

	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
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

	createInputChannelOptions := core.CreateChannelOptions{}
	createOutputChannelOptions := core.CreateChannelOptions{}
	createFunctionOptions := core.CreateFunctionOptions{}
	createSubscriptionOptions := core.CreateSubscriptionOptions{}

	invokers := map[string]string{
		"command": "https://github.com/projectriff/command-function-invoker/raw/v0.0.7/command-invoker.yaml",
		"java":    "https://github.com/projectriff/java-function-invoker/raw/v0.0.7/java-invoker.yaml",
		"node":    "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
	}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a new function resource, with optional input and output channels",
		Long: `Create a new function resource from the content of the provided Git repo/revision.

The INVOKER arg defines the language invoker that is added to the function code in the build step. The resulting image is 
then used to create a Knative Service (service.serving.knative.dev) instance of the name specified for the function. 
From then on you can use the sub-commands for the 'service' command to interact with the service created for the function. 

` + channelLongDesc + `

` + envFromLongDesc + `
`,
		Example: `  riff function create node square --git-repo https://github.com/acme/square --image acme/square --namespace joseph-ns
  riff function create java tweets-logger --git-repo https://github.com/acme/tweets --image acme/tweets-logger:1.0.0 --input tweets --bus kafka`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionCreateNumberOfArgs),
			AtPosition(functionCreateInvokerIndex, ValidName()),
			AtPosition(functionCreateFunctionNameIndex, ValidName()),
		),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsValidationConjunction(
				FlagsDependency(Set("input"), exactlyOneOfBusOrClusterBus),
				FlagsDependency(NotSet("input"), NoneOf("bus", "cluster-bus")),
				FlagsDependency(NotSet("input"), NoneOf("output")),
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			fnName := args[functionCreateFunctionNameIndex]
			invoker := args[functionCreateInvokerIndex]
			invokerURL, exists := invokers[invoker]
			if !exists {
				return fmt.Errorf("unknown invoker: %s", invoker)
			}

			createFunctionOptions.Name = fnName
			createFunctionOptions.InvokerURL = invokerURL
			f, err := (*fcTool).CreateFunction(createFunctionOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			var c *v1alpha1.Channel
			var subscr *v1alpha1.Subscription
			if createInputChannelOptions.Name != "" {
				c, err = (*fcTool).CreateChannel(createInputChannelOptions)
				if err != nil {
					return err
				}

				if createOutputChannelOptions.Name != "" {
					c, err = (*fcTool).CreateChannel(createOutputChannelOptions)
					if err != nil {
						return err
					}
				}
				createSubscriptionOptions.Name = subscriptionNameFromService(fnName)
				createSubscriptionOptions.Subscriber = subscriberNameFromService(fnName)
				subscr, err = (*fcTool).CreateSubscription(createSubscriptionOptions)
				if err != nil {
					return err
				}
			}

			if createFunctionOptions.DryRun {
				marshaller := NewMarshaller(cmd.OutOrStdout())
				if err = marshaller.Marshal(f); err != nil {
					return err
				}
				if c != nil {
					if err = marshaller.Marshal(c); err != nil {
						return err
					}
				}
				if subscr != nil {
					if err = marshaller.Marshal(subscr); err != nil {
						return err
					}
				}
			} else {
				printSuccessfulCompletion(cmd)
				if !createFunctionOptions.Verbose && !createFunctionOptions.Wait {
					fmt.Fprintf(cmd.OutOrStdout(), "Issue `riff service status %s` to see the status of the function\n", fnName)
				}
			}

			return nil
		},
	}

	LabelArgs(command, "INVOKER", "FUNCTION_NAME")

	command.Flags().VarP(
		BroadcastStringValue("",
			&createFunctionOptions.Namespace,
			&createInputChannelOptions.Namespace,
			&createOutputChannelOptions.Namespace,
			&createSubscriptionOptions.Namespace,
		),
		"namespace", "n", "the `namespace` of the subscription, channel, and function",
	)

	command.Flags().VarP(
		BroadcastStringValue("",
			&createInputChannelOptions.Name,
			&createSubscriptionOptions.Channel,
		),
		"input", "i", "name of the function's input `channel`, if any",
	)

	command.Flags().VarP(
		BroadcastStringValue("",
			&createOutputChannelOptions.Name,
			&createSubscriptionOptions.ReplyTo,
		),
		"output", "o", "name of the function's output `channel`, if any",
	)

	command.Flags().VarPF(
		BroadcastBoolValue(false,
			&createFunctionOptions.DryRun,
			&createInputChannelOptions.DryRun,
			&createOutputChannelOptions.DryRun,
			&createSubscriptionOptions.DryRun,
		),
		"dry-run", "", dryRunUsage,
	).NoOptDefVal = "true"

	command.Flags().Var(
		BroadcastStringValue("",
			&createInputChannelOptions.Bus,
			&createOutputChannelOptions.Bus,
		),
		"bus", busUsage,
	)

	command.Flags().Var(
		BroadcastStringValue("",
			&createInputChannelOptions.ClusterBus,
			&createOutputChannelOptions.ClusterBus,
		),
		"cluster-bus", clusterBusUsage,
	)

	command.Flags().StringVar(&createFunctionOptions.Image, "image", "", "the name of the image to build; must be a writable `repository/image[:tag]` with credentials configured")
	command.MarkFlagRequired("image")
	command.Flags().StringVar(&createFunctionOptions.GitRepo, "git-repo", "", "the `URL` for a git repository hosting the function code")
	command.MarkFlagRequired("git-repo")
	command.Flags().StringVar(&createFunctionOptions.GitRevision, "git-revision", "master", "the git `ref-spec` of the function code to use")
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

			printSuccessfulCompletion(cmd)

			return nil
		},
	}

	LabelArgs(command, "FUNCTION_NAME")

	command.Flags().StringVarP(&buildFunctionOptions.Namespace, "namespace", "n", "", "the `namespace` of the function")
	command.Flags().BoolVarP(&buildFunctionOptions.Verbose, "verbose", "v", false, verboseUsage)
	command.Flags().BoolVarP(&buildFunctionOptions.Wait, "wait", "w", false, waitUsage)

	return command
}
