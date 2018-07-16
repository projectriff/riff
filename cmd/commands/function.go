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
	"github.com/pivotal-cf-experimental/riff-cli/pkg/tool"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	functionCreateInvokerIndex = iota
	functionCreateFunctionIndex
	functionCreateNumberOfArgs
)

const (
	functionSubscribeFunctionIndex = iota
	functionSubscribeNumberOfArgs
)

const (
	functionDeleteFunctionIndex = iota
	functionDeleteNumberOfArgs
)

func Function() *cobra.Command {
	return &cobra.Command{
		Use:   "function",
		Short: "interact with function related resources",
	}
}

func FunctionCreate(fcTool *tool.Client) *cobra.Command {

	var exactlyOneOfImageOrBuild = FlagsValidationConjunction(
		AtLeastOneOf("image", "build"),
		AtMostOneOf("image", "build"),
	)

	createChannelOptions := tool.CreateChannelOptions{}
	createFunctionOptions := tool.CreateFunctionOptions{}
	createSubscriptionOptions := tool.CreateSubscriptionOptions{}

	command := &cobra.Command{
		Use:   "create",
		Short: "create a new function resource, with optional input binding",
		Example: `  riff function create node square --image acme/square:1.0 --namespace joseph-ns
  riff function create java tweets-logger --image acme/tweets-logger:1.0.0 --input tweets --bus kafka`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionCreateNumberOfArgs),
			AtPosition(functionCreateInvokerIndex, ValidName()),
			AtPosition(functionCreateFunctionIndex, ValidName()),
		),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsValidationConjunction(
				FlagsDependency("input", exactlyOneOfBusOrClusterBus),
				exactlyOneOfImageOrBuild,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			e := yaml.NewEncoder(cmd.OutOrStdout())

			fnName := args[functionCreateFunctionIndex]
			createFunctionOptions.Name = fnName
			f, err := (*fcTool).CreateFunction(createFunctionOptions)
			if err != nil {
				return err
			}
			if err = e.Encode(f); err != nil {
				return err
			}

			if createChannelOptions.Name != "" {
				c, err := (*fcTool).CreateChannel(createChannelOptions)
				if err != nil {
					return err
				}
				if err = e.Encode(c); err != nil {
					return err
				}

				createSubscriptionOptions.Name = subscriptionNameFromFunction(fnName)
				createSubscriptionOptions.Subscriber = subscriberNameFromFunction(fnName) // TODO
				subscr, err := (*fcTool).CreateSubscription(createSubscriptionOptions)
				if err != nil {
					return err
				}
				if err = e.Encode(subscr); err != nil {
					return err
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

	command.Flags().StringVar(&createFunctionOptions.Image, "image", "", "reference to an already built `name[:tag]` image that contains the function.")
	command.Flags().String("build", "", "TODO: build options?")

	return command
}

func FunctionSubscribe(fcClient *tool.Client) *cobra.Command {

	createSubscriptionOptions := tool.CreateSubscriptionOptions{}

	command := &cobra.Command{
		Use:     "subscribe",
		Short:   "subscribe a function to an existing input channel",
		Example: `  riff function subscribe square --input numbers --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionSubscribeNumberOfArgs),
			AtPosition(functionSubscribeFunctionIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			e := yaml.NewEncoder(cmd.OutOrStdout())

			fnName := args[functionSubscribeFunctionIndex]

			createSubscriptionOptions.Name = subscriptionNameFromFunction(fnName)
			createSubscriptionOptions.Subscriber = subscriberNameFromFunction(fnName)
			s, err := (*fcClient).CreateSubscription(createSubscriptionOptions)
			if err != nil {
				return err
			}
			if err = e.Encode(s); err != nil {
				return err
			}

			return err
		},
	}

	LabelArgs(command, "<function-name>")

	command.Flags().StringVarP(&createSubscriptionOptions.Channel, "input", "i", "", "name of the input `channel` to subscribe the function to.")
	command.MarkFlagRequired("input")
	command.Flags().StringVarP(&createSubscriptionOptions.Namespace, "namespace", "n", "", namespaceUsage)

	return command
}

func FunctionDelete(fcClient *tool.Client) *cobra.Command {

	deleteFunctionOptions := tool.DeleteFunctionOptions{}

	command := &cobra.Command{
		Use:     "delete",
		Short:   "delete an existing function",
		Example: `  riff function delete square --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(functionDeleteNumberOfArgs),
			AtPosition(functionDeleteFunctionIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[functionDeleteFunctionIndex]
			deleteFunctionOptions.Name = fnName
			return (*fcClient).DeleteFunction(deleteFunctionOptions)
		},
	}

	LabelArgs(command, "<function-name>")

	command.Flags().StringVarP(&deleteFunctionOptions.Namespace, "namespace", "n", "", namespaceUsage)

	return command
}

// TODO
func subscriberNameFromFunction(fnName string) string {
	return fnName
}

func subscriptionNameFromFunction(fnName string) string {
	return fnName
}
