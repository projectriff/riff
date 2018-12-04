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
	"os"
	"os/exec"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/frioux/shellquote"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
)

const (
	serviceCreateServiceNameIndex = iota
	serviceCreateNumberOfArgs
)

const (
	serviceUpdateServiceNameIndex = iota
	serviceUpdateNumberOfArgs
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
	serviceInvokeServicePathIndex
	serviceInvokeMaxNumberOfArgs
)

const (
	serviceDeleteServiceNameIndex = iota
	serviceDeleteNumberOfArgs
)

func Service() *cobra.Command {
	return &cobra.Command{
		Use:   "service",
		Short: "Interact with service related resources",
		Long:  "Interact with service (as in `service.serving.knative.dev`) related resources.",
	}
}

func ServiceCreate(fcTool *core.Client) *cobra.Command {

	createServiceOptions := core.CreateOrUpdateServiceOptions{}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a new service resource",
		Long: `Create a new service resource from a given image.

` + envFromLongDesc + `
`,
		Example: `  ` + env.Cli.Name + ` service create square --image acme/square:1.0 --namespace joseph-ns
  ` + env.Cli.Name + ` service create greeter --image acme/greeter:1.0 --env FOO=bar --env MESSAGE=Hello
  ` + env.Cli.Name + ` service create tweets-logger --image acme/tweets-logger:1.0.0`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(serviceCreateNumberOfArgs),
			AtPosition(serviceCreateServiceNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			fnName := args[serviceCreateServiceNameIndex]
			createServiceOptions.Name = fnName
			f, err := (*fcTool).CreateService(createServiceOptions)
			if err != nil {
				return err
			}

			if createServiceOptions.DryRun {
				marshaller := NewMarshaller(cmd.OutOrStdout())
				if err = marshaller.Marshal(f); err != nil {
					return err
				}
			} else {
				PrintSuccessfulCompletion(cmd)
			}

			return nil
		},
	}

	LabelArgs(command, "SERVICE_NAME")

	command.Flags().StringVarP(&createServiceOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")
	command.Flags().BoolVarP(&createServiceOptions.DryRun, "dry-run", "", false, dryRunUsage)

	command.Flags().StringVar(&createServiceOptions.Image, "image", "", "the `name[:tag]` reference of an image containing the application/function")
	command.MarkFlagRequired("image")

	command.Flags().StringArrayVar(&createServiceOptions.Env, "env", []string{}, envUsage)
	command.Flags().StringArrayVar(&createServiceOptions.EnvFrom, "env-from", []string{}, envFromUsage)

	return command
}

func ServiceUpdate(client *core.Client) *cobra.Command {
	updateServiceOptions := core.CreateOrUpdateServiceOptions{}

	command := &cobra.Command{
		Use:     "update",
		Short:   "Create a new revision for a service, with updated attributes",
		Long:    `Create a new revision for a service, updating the application/function image and/or environment.`,
		Example: `  ` + env.Cli.Name + ` service update square --image acme/square:1.1 --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(serviceUpdateNumberOfArgs),
			AtPosition(serviceUpdateServiceNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[serviceUpdateServiceNameIndex]
			updateServiceOptions.Name = fnName
			svc, err := (*client).UpdateService(updateServiceOptions)
			if err != nil {
				return err
			}
			if updateServiceOptions.DryRun {
				marshaller := NewMarshaller(cmd.OutOrStdout())
				if err = marshaller.Marshal(svc); err != nil {
					return err
				}
			} else {
				PrintSuccessfulCompletion(cmd)
			}

			return nil
		},
	}

	LabelArgs(command, "SERVICE_NAME")

	command.Flags().StringVarP(&updateServiceOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")
	command.Flags().BoolVar(&updateServiceOptions.DryRun, "dry-run", false, dryRunUsage)
	command.Flags().StringVar(&updateServiceOptions.Image, "image", "", "the `name[:tag]` reference of an image containing the application/function")
	command.Flags().StringArrayVar(&updateServiceOptions.Env, "env", []string{}, envUsage)
	command.Flags().StringArrayVar(&updateServiceOptions.EnvFrom, "env-from", []string{}, envFromUsage)

	return command
}

func ServiceStatus(fcClient *core.Client) *cobra.Command {

	serviceStatusOptions := core.ServiceStatusOptions{}

	command := &cobra.Command{
		Use:     "status",
		Short:   "Display the status of a service",
		Example: `  ` + env.Cli.Name + ` service status square --namespace joseph-ns`,
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

			fmt.Fprintf(cmd.OutOrStdout(), "Last Transition Time:        %s\n", cond.LastTransitionTime.Inner.Format(time.RFC3339))

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
		Example: `  ` + env.Cli.Name + ` service list
  ` + env.Cli.Name + ` service list --namespace joseph-ns`,
		Args: cobra.ExactArgs(serviceListNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := (*fcClient).ListServices(listServiceOptions)
			if err != nil {
				return err
			}

			Display(cmd.OutOrStdout(), serviceToInterfaceSlice(services.Items), makeServiceExtractors())
			PrintSuccessfulCompletion(cmd)
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
		Example: `  ` + env.Cli.Name + ` service invoke square --namespace joseph-ns
  ` + env.Cli.Name + ` service invoke square /foo -- --data 42`,
		PreRunE: FlagsValidatorAsCobraRunE(AtMostOneOf("json", "text")),
		Args: UpToDashDash(ArgValidationConjunction(
			cobra.MinimumNArgs(serviceInvokeMaxNumberOfArgs-1),
			cobra.MaximumNArgs(serviceInvokeMaxNumberOfArgs),
			AtPosition(serviceInvokeServiceNameIndex, ValidName()),
		)),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsLengthAtDash := cmd.ArgsLenAtDash()
			serviceInvokeOptions.Name = args[serviceInvokeServiceNameIndex]
			path := "/"
			if argsLengthAtDash > serviceInvokeServicePathIndex ||
				argsLengthAtDash == -1 && len(args) > serviceInvokeServicePathIndex {
				path = args[serviceInvokeServicePathIndex]
			}
			ingress, hostName, err := (*fcClient).ServiceCoordinates(serviceInvokeOptions)
			if err != nil {
				return err
			}

			curlCmd := exec.Command("curl", ingress+path)

			curlCmd.Stdin = os.Stdin
			curlCmd.Stdout = cmd.OutOrStdout()
			curlCmd.Stderr = cmd.OutOrStderr()

			hostHeader := fmt.Sprintf("Host: %s", hostName)
			curlCmd.Args = append(curlCmd.Args, "-H", hostHeader)

			if serviceInvokeOptions.ContentTypeJson {
				curlCmd.Args = append(curlCmd.Args, "-H", "Content-Type: application/json")
			} else if serviceInvokeOptions.ContentTypeText {
				curlCmd.Args = append(curlCmd.Args, "-H", "Content-Type: text/plain")
			}

			if argsLengthAtDash > 0 {
				curlCmd.Args = append(curlCmd.Args, args[argsLengthAtDash:]...)
			}

			quoted, err := shellquote.Quote(curlCmd.Args)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), quoted)

			return curlCmd.Run()
		},
	}

	LabelArgs(command, "SERVICE_NAME", "PATH")

	command.Flags().StringVarP(&serviceInvokeOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")
	command.Flags().BoolVar(&serviceInvokeOptions.ContentTypeJson, "json", false, "set the request's content type to 'application/json'")
	command.Flags().BoolVar(&serviceInvokeOptions.ContentTypeText, "text", false, "set the request's content type to 'text/plain'")

	return command
}

func ServiceDelete(fcClient *core.Client) *cobra.Command {

	deleteServiceOptions := core.DeleteServiceOptions{}

	command := &cobra.Command{
		Use:     "delete",
		Short:   "Delete an existing service",
		Example: `  ` + env.Cli.Name + ` service delete square --namespace joseph-ns`,
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

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "SERVICE_NAME")

	command.Flags().StringVarP(&deleteServiceOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")

	return command
}

func serviceToInterfaceSlice(subscriptions []servingv1alpha1.Service) []interface{} {
	result := make([]interface{}, len(subscriptions))
	for i := range subscriptions {
		result[i] = subscriptions[i]
	}
	return result
}

func makeServiceExtractors() []NamedExtractor {
	return []NamedExtractor{
		{
			name: "NAME",
			fn:   func(s interface{}) string { return s.(servingv1alpha1.Service).Name },
		},
		{
			name: "STATUS",
			fn: func(s interface{}) string {
				service := s.(servingv1alpha1.Service)
				cond := service.Status.GetCondition(servingv1alpha1.ServiceConditionReady)
				if cond == nil {
					return "Unknown"
				}
				switch cond.Status {
				case v1.ConditionTrue:
					return "Running"
				case v1.ConditionFalse:
					return fmt.Sprintf("%s: %s", cond.Reason, cond.Message)
				default:
					return "Unknown"
				}
			},
		},
	}
}
