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
	"github.com/projectriff/riff-cli/pkg/core"
	"github.com/spf13/cobra"
)

const (
	functionCreateInvokerIndex = iota
	functionCreateFunctionNameIndex
	functionCreateNumberOfArgs
)

func Function() *cobra.Command {
	return &cobra.Command{
		Use:   "function",
		Short: "interact with function related resources",
	}
}

func FunctionCreate(fcTool *core.Client) *cobra.Command {

	createChannelOptions := core.CreateChannelOptions{}
	createFunctionOptions := core.CreateFunctionOptions{}
	createSubscriptionOptions := core.CreateSubscriptionOptions{}

	var write, force = false, false

	command := &cobra.Command{
		Use:   "create",
		Short: "create a new function resource, with optional input binding",
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
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			fnName := args[functionCreateFunctionNameIndex]
			invoker := args[functionCreateInvokerIndex]
			invokerURL := fmt.Sprintf("https://github.com/projectriff/%s-function-invoker/raw/v0.0.7/%s-invoker.yaml", invoker, invoker)

			createFunctionOptions.Name = fnName
			createFunctionOptions.InvokerURL = invokerURL
			f, err := (*fcTool).CreateFunction(createFunctionOptions)
			if err != nil {
				return err
			}

			var c *v1alpha1.Channel
			var subscr *v1alpha1.Subscription
			if createChannelOptions.Name != "" {
				c, err = (*fcTool).CreateChannel(createChannelOptions)
				if err != nil {
					return err
				}

				createSubscriptionOptions.Name = subscriptionNameFromService(fnName)
				createSubscriptionOptions.Subscriber = subscriberNameFromService(fnName)
				subscr, err = (*fcTool).CreateSubscription(createSubscriptionOptions)
				if err != nil {
					return err
				}
			}

			if write {
				fmarshaller, err := NewMarshaller(fmt.Sprintf("%s-service.yaml", fnName), force)
				if err != nil {
					return err
				}
				if err = fmarshaller.Marshal(f); err != nil {
					return err
				}
				if createChannelOptions.Name != "" {
					cmarshaller, err := NewMarshaller(fmt.Sprintf("%s-channel.yaml", createChannelOptions.Name), force)
					if err = cmarshaller.Marshal(c); err != nil {
						return err
					}
					smarshaller, err := NewMarshaller(fmt.Sprintf("%s-subscription.yaml", subscr.Name), force)
					if err = smarshaller.Marshal(subscr); err != nil {
						return err
					}

				}
			}

			return err
		},
	}

	LabelArgs(command, "<invoker>", "<function-name>")

	command.Flags().VarP(
		BroadcastStringValue("",
			&createFunctionOptions.Namespace,
			&createChannelOptions.Namespace,
			&createSubscriptionOptions.Namespace,
		),
		"namespace", "n", namespaceUsage,
	)

	command.Flags().VarP(
		BroadcastStringValue("",
			&createChannelOptions.Name,
			&createSubscriptionOptions.Channel,
		),
		"input", "i", "name of the input `channel` to subscribe the function to.",
	)

	command.Flags().StringVar(&createChannelOptions.Bus, "bus", "", busUsage)
	command.Flags().StringVar(&createChannelOptions.ClusterBus, "cluster-bus", "", clusterBusUsage)

	command.Flags().StringVar(&createFunctionOptions.Image, "image", "", "the name of the image to build. Must be a writable `repository/image[:tag]` with write credentials configured.")
	command.MarkFlagRequired("image")
	command.Flags().StringVar(&createFunctionOptions.GitRepo, "git-repo", "", "the `URL` for the git repo hosting the function source.")
	command.MarkFlagRequired("git-repo")
	command.Flags().StringVar(&createFunctionOptions.GitRevision, "git-revision", "master", "the git `ref-spec` to build.")
	command.Flags().StringVar(&createFunctionOptions.Handler, "handler", "", "name of `method or class` to invoke. See specific invoker for detail.")
	command.Flags().StringVar(&createFunctionOptions.Artifact, "artifact", "", "`path` to the function artifact, source code or jar file. Attempts detection if not specified.")

	command.Flags().BoolVarP(&write, "write", "w", false, "whether to write yaml files for created resources.")
	command.Flags().BoolVarP(&force, "force", "f", false, "force writing of files if they already exist.")

	return command
}
