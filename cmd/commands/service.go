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

	"os"
	"os/exec"

	"github.com/frioux/shellquote"
	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	v1alpha12 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
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
	serviceListNumberOfArgs = iota
)

const (
	serviceInvokeServiceNameIndex = iota
	serviceInvokeNumberOfArgs
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
		Short: "Interact with service related resources",
		Long:  "Interact with service (as in service.serving.knative.dev) related resources.",
	}
}

func ServiceCreate(fcTool *core.Client) *cobra.Command {

	createChannelOptions := core.CreateChannelOptions{}
	createServiceOptions := core.CreateServiceOptions{}
	createSubscriptionOptions := core.CreateSubscriptionOptions{}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a new service resource, with optional input binding",
		Long: `Create a new service resource from a given image.
If an input channel and bus are specified, create the channel in the bus and subscribe the service to the channel.`,
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

			if createServiceOptions.DryRun {
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
			}

			return nil
		},
	}

	LabelArgs(command, "SERVICE_NAME")

	command.Flags().VarP(
		BroadcastStringValue("",
			&createServiceOptions.Namespace,
			&createChannelOptions.Namespace,
			&createSubscriptionOptions.Namespace,
		),
		"namespace", "n", "the `namespace` of the service and any namespaced resources specified",
	)

	command.Flags().VarP(
		BroadcastStringValue("",
			&createChannelOptions.Name,
			&createSubscriptionOptions.Channel,
		),
		"input", "i", "name of the service's input `channel`, if any",
	)

	command.Flags().VarPF(
		BroadcastBoolValue(false,
			&createServiceOptions.DryRun,
			&createChannelOptions.DryRun,
			&createSubscriptionOptions.DryRun,
		),
		"dry-run", "", dryRunUsage,
	).NoOptDefVal = "true"

	command.Flags().StringVar(&createChannelOptions.Bus, "bus", "", busUsage)
	command.Flags().StringVar(&createChannelOptions.ClusterBus, "cluster-bus", "", clusterBusUsage)

	command.Flags().StringVar(&createServiceOptions.Image, "image", "", "the `name[:tag]` reference of an image containing the application/function")
	command.MarkFlagRequired("image")
	return command
}

func ServiceStatus(fcClient *core.Client) *cobra.Command {

	serviceStatusOptions := core.ServiceStatusOptions{}

	command := &cobra.Command{
		Use:     "status",
		Short:   "Display the status of a service",
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

			fmt.Fprintf(cmd.OutOrStdout(), "Last Transition Time:        %s\n", cond.LastTransitionTime.Format(time.RFC3339))

			if cond.Reason != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Message:                     %s\n", cond.Message)
				fmt.Fprintf(cmd.OutOrStdout(), "Reason:                      %s\n", cond.Reason)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Status:                      %s\n", cond.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Type:                        %s\n", cond.Type)

			return nil
		},
	}

	LabelArgs(command, "SERVICE_NAME")

	command.Flags().StringVarP(&serviceStatusOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")

	return command
}

func ServiceList(fcClient *core.Client) *cobra.Command {
	listServiceOptions := core.ListServiceOptions{}

	command := &cobra.Command{
		Use:   "list",
		Short: "List service resources",
		Example: `  riff service list
  riff service list --namespace joseph-ns`,
		Args: cobra.ExactArgs(serviceListNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := (*fcClient).ListServices(listServiceOptions)
			if err != nil {
				return err
			}

			if len(services.Items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No resources found.")
			} else {
				maxServiceNameLength := len("NAME ") //Make sure column names have enough room, even with short service names
				for _, service := range services.Items {
					if len(service.Name) > maxServiceNameLength {
						maxServiceNameLength = len(service.Name)
					}
				}
				pad := fmt.Sprintf("%%-%ds%%s\n", maxServiceNameLength+1)

				fmt.Fprintf(cmd.OutOrStdout(), pad, "NAME", "STATUS")
				for _, service := range services.Items {
					cond := service.Status.GetCondition(v1alpha12.ServiceConditionReady)
					var status string
					if cond == nil {
						status = "Unknown"
					} else {
						switch cond.Status {
						case v1.ConditionTrue:
							status = "Running"
						case v1.ConditionFalse:
							status = fmt.Sprintf("%s: %s", cond.Reason, cond.Message)
						default:
							status = "Unknown"
						}
					}

					fmt.Fprintf(cmd.OutOrStdout(), pad, service.Name, status)
				}
			}

			return nil
		},
	}

	command.Flags().StringVarP(&listServiceOptions.Namespace, "namespace", "n", "", "the `namespace` of the services to be listed")

	return command
}

func ServiceInvoke(fcClient *core.Client) *cobra.Command {

	serviceInvokeOptions := core.ServiceInvokeOptions{}

	command := &cobra.Command{
		Use:   "invoke",
		Short: "Invoke a service",
		Long: `Invoke a service by shelling out to curl.

The curl command is printed so it can be copied and extended.

Additional curl arguments and flags may be specified after a double dash (--).`,
		Example: `  riff service invoke square --namespace joseph-ns
  riff service invoke square -- --include`,
		Args: UpToDashDash(ArgValidationConjunction(
			cobra.ExactArgs(serviceInvokeNumberOfArgs),
			AtPosition(serviceInvokeServiceNameIndex, ValidName()),
		)),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceInvokeOptions.Name = args[serviceInvokeServiceNameIndex]
			ingress, hostName, err := (*fcClient).ServiceCoordinates(serviceInvokeOptions)
			if err != nil {
				return err
			}

			curlCmd := exec.Command("curl", ingress)

			curlCmd.Stdin = os.Stdin
			curlCmd.Stdout = cmd.OutOrStdout()
			curlCmd.Stderr = cmd.OutOrStderr()

			hostHeader := fmt.Sprintf("Host: %s", hostName)
			curlCmd.Args = append(curlCmd.Args, "-H", hostHeader)

			if cmd.ArgsLenAtDash() > 0 {
				curlCmd.Args = append(curlCmd.Args, args[cmd.ArgsLenAtDash():]...)
			}

			quoted, err := shellquote.Quote(curlCmd.Args)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), quoted)

			return curlCmd.Run()
		},
	}

	LabelArgs(command, "SERVICE_NAME")

	command.Flags().StringVarP(&serviceInvokeOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")

	return command
}

func ServiceSubscribe(fcClient *core.Client) *cobra.Command {

	createSubscriptionOptions := core.CreateSubscriptionOptions{}

	command := &cobra.Command{
		Use:     "subscribe",
		Short:   "Subscribe a service to an existing input channel",
		Example: `  riff service subscribe square --input numbers --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(serviceSubscribeNumberOfArgs),
			AtPosition(serviceSubscribeServiceNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			fnName := args[serviceSubscribeServiceNameIndex]

			if createSubscriptionOptions.Name == "" {
				createSubscriptionOptions.Name = subscriptionNameFromService(fnName)
			}
			createSubscriptionOptions.Subscriber = subscriberNameFromService(fnName)
			s, err := (*fcClient).CreateSubscription(createSubscriptionOptions)
			if err != nil {
				return err
			}
			if createSubscriptionOptions.DryRun {
				marshaller := NewMarshaller(cmd.OutOrStdout())
				if err = marshaller.Marshal(s); err != nil {
					return err
				}
			} else {
				printSuccessfulCompletion(cmd)
			}
			return nil
		},
	}

	LabelArgs(command, "SERVICE_NAME")

	command.Flags().StringVar(&createSubscriptionOptions.Name, "subscription", "", "`name` of the subscription (default SERVICE_NAME)")
	command.Flags().StringVarP(&createSubscriptionOptions.Channel, "input", "i", "", "the name of an input `channel` for the service")
	command.MarkFlagRequired("input")
	command.Flags().StringVarP(&createSubscriptionOptions.Namespace, "namespace", "n", "", "the `namespace` of the subscription, channel, and service")
	command.Flags().BoolVar(&createSubscriptionOptions.DryRun, "dry-run", false, dryRunUsage)

	return command
}

func ServiceDelete(fcClient *core.Client) *cobra.Command {

	deleteServiceOptions := core.DeleteServiceOptions{}

	command := &cobra.Command{
		Use:     "delete",
		Short:   "Delete an existing service",
		Example: `  riff service delete square --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(serviceDeleteNumberOfArgs),
			AtPosition(serviceDeleteServiceNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[serviceDeleteServiceNameIndex]
			deleteServiceOptions.Name = fnName
			err := (*fcClient).DeleteService(deleteServiceOptions)
			if err != nil {
				return err
			}

			printSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "SERVICE_NAME")

	command.Flags().StringVarP(&deleteServiceOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")

	return command
}

// subscriptionNameFromService returns the name to use for the subscription being created alongside
// a service/function. By convention, this is chosen to be the name of the service.
func subscriptionNameFromService(fnName string) string {
	return fnName
}

// subscriberNameFromService returns the name to use for the `subscriber` field of a subscription,
// given a service/function that is being created/subscribed. This has to be the name of the service itself.
func subscriberNameFromService(fnName string) string {
	return fnName
}
