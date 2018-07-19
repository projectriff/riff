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

	"time"

	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	"github.com/projectriff/riff-cli/pkg/core"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	serviceCreateServiceNameIndex = iota
	serviceCreateNumberOfArgs
)

const (
	serviceStatusServiceNameIndex = iota
	serviceStatusNumberOfArgs
)

const (
	serviceSubscribeServiceNameIndex = iota
	serviceSubscribeNumberOfArgs
)

const (
	serviceDeleteServiceNameIndex = iota
	serviceDeleteNumberOfArgs
)

func Service() *cobra.Command {
	return &cobra.Command{
		Use:   "service",
		Short: "interact with service related resources",
	}
}

func ServiceCreate(fcTool *core.Client) *cobra.Command {

	createChannelOptions := core.CreateChannelOptions{}
	createServiceOptions := core.CreateServiceOptions{}
	createSubscriptionOptions := core.CreateSubscriptionOptions{}

	var write, force = false, false

	command := &cobra.Command{
		Use:   "create",
		Short: "create a new service resource, with optional input binding",
		Example: `  riff service create square --image acme/square:1.0 --namespace joseph-ns
  riff service create tweets-logger --image acme/tweets-logger:1.0.0 --input tweets --bus kafka`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(serviceCreateNumberOfArgs),
			AtPosition(serviceCreateServiceNameIndex, ValidName()),
		),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsValidationConjunction(
				FlagsDependency(Set("input"), exactlyOneOfBusOrClusterBus),
				FlagsDependency(NotSet("input"), NoneOf("bus", "cluster-bus")),
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			fnName := args[serviceCreateServiceNameIndex]
			createServiceOptions.Name = fnName
			f, err := (*fcTool).CreateService(createServiceOptions)
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
				createSubscriptionOptions.Subscriber = subscriberNameFromService(fnName) // TODO
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

	LabelArgs(command, "<service-name>")

	command.Flags().VarP(
		BroadcastStringValue("",
			&createServiceOptions.Namespace,
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
		"input", "i", "name of the input `channel` to subscribe the service to.",
	)

	command.Flags().StringVar(&createChannelOptions.Bus, "bus", "", busUsage)
	command.Flags().StringVar(&createChannelOptions.ClusterBus, "cluster-bus", "", clusterBusUsage)

	command.Flags().StringVar(&createServiceOptions.Image, "image", "", "reference to an already built `name[:tag]` image that contains the application/function.")

	command.Flags().BoolVarP(&write, "write", "w", false, "whether to write yaml files for created resources.")
	command.Flags().BoolVarP(&force, "force", "f", false, "force writing of files if they already exist.")

	return command
}

func ServiceStatus(fcClient *core.Client) *cobra.Command {

	serviceStatusOptions := core.ServiceStatusOptions{}

	command := &cobra.Command{
		Use:     "status",
		Short:   "display the status of a service",
		Example: `  riff service status square --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(serviceStatusNumberOfArgs),
			AtPosition(serviceStatusServiceNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[serviceStatusServiceNameIndex]
			serviceStatusOptions.Name = fnName
			cond, err := (*fcClient).ServiceStatus(serviceStatusOptions)
			if err != nil {
				return err
			}

			fmt.Printf("Last Transition Time:        %s\n", cond.LastTransitionTime.Format(time.RFC3339))

			if cond.Reason != "" {
				fmt.Printf("Message:                     %s\n", cond.Message)
				fmt.Printf("Reason:                      %s\n", cond.Reason)
			}

			fmt.Printf("Status:                      %s\n", cond.Status)
			fmt.Printf("Type:                        %s\n", cond.Type)

			return nil
		},
	}

	LabelArgs(command, "<function-name>")

	command.Flags().StringVarP(&serviceStatusOptions.Namespace, "namespace", "n", "", namespaceUsage)

	return command
}

func ServiceSubscribe(fcClient *core.Client) *cobra.Command {

	createSubscriptionOptions := core.CreateSubscriptionOptions{}

	command := &cobra.Command{
		Use:     "subscribe",
		Short:   "subscribe a service to an existing input channel",
		Example: `  riff service subscribe square --input numbers --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(serviceSubscribeNumberOfArgs),
			AtPosition(serviceSubscribeServiceNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			e := yaml.NewEncoder(cmd.OutOrStdout())

			fnName := args[serviceSubscribeServiceNameIndex]

			createSubscriptionOptions.Name = subscriptionNameFromService(fnName)
			createSubscriptionOptions.Subscriber = subscriberNameFromService(fnName)
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

	LabelArgs(command, "<service-name>")

	command.Flags().StringVarP(&createSubscriptionOptions.Channel, "input", "i", "", "name of the input `channel` to subscribe the service to.")
	command.MarkFlagRequired("input")
	command.Flags().StringVarP(&createSubscriptionOptions.Namespace, "namespace", "n", "", namespaceUsage)

	return command
}

func ServiceDelete(fcClient *core.Client) *cobra.Command {

	deleteServiceOptions := core.DeleteServiceOptions{}

	command := &cobra.Command{
		Use:     "delete",
		Short:   "delete an existing service",
		Example: `  riff service delete square --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(serviceDeleteNumberOfArgs),
			AtPosition(serviceDeleteServiceNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[serviceDeleteServiceNameIndex]
			deleteServiceOptions.Name = fnName
			return (*fcClient).DeleteService(deleteServiceOptions)
		},
	}

	LabelArgs(command, "<service-name>")

	command.Flags().StringVarP(&deleteServiceOptions.Namespace, "namespace", "n", "", namespaceUsage)

	return command
}

// TODO
func subscriberNameFromService(fnName string) string {
	return fnName
}

func subscriptionNameFromService(fnName string) string {
	return fnName
}
