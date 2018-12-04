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
	"io"
	"strings"
	"text/template"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/env"
	. "github.com/spf13/cobra"
)

const (
	subscriptionCreateNameIndex = iota
	subscriptionCreateMaxNumberOfArgs
)
const (
	subscriptionDeleteNameIndex = iota
	subscriptionDeleteNumberOfArgs
)

const (
	subscriptionListNumberOfArgs = iota
)

func Subscription() *Command {
	return &Command{
		Use:   "subscription",
		Short: "Interact with subscription-related resources",
	}
}

func SubscriptionCreate(client *core.Client) *Command {
	options := core.CreateSubscriptionOptions{}

	command := &Command{
		Use:   "create",
		Short: "Create a new subscription, binding a service to an input channel",
		Long: "Create a new, optionally named subscription, binding a service to an input channel. " +
			"The default name of the subscription is the provided subscriber name. " +
			"The subscription can optionally be bound to an output channel.",
		Example: `  ` + env.Cli.Name + ` subscription create --channel tweets --subscriber tweets-logger
  ` + env.Cli.Name + ` subscription create my-subscription --channel tweets --subscriber tweets-logger
  ` + env.Cli.Name + ` subscription create --channel tweets --subscriber tweets-logger --reply logged-tweets`,
		Args: ArgValidationConjunction(
			MaximumNArgs(subscriptionCreateMaxNumberOfArgs),
			OptionalAtPosition(subscriptionCreateNameIndex, ValidName()),
		),
		RunE: func(cmd *Command, args []string) error {
			options.Name = computeSubscriptionName(args, options)
			_, err := (*client).CreateSubscription(options)
			if err != nil {
				return err
			}
			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "SUBSCRIPTION_NAME")
	defineFlagsForCreate(command, &options)
	return command
}

func SubscriptionDelete(client *core.Client) *Command {
	deleteOptions := core.DeleteSubscriptionOptions{}

	command := &Command{
		Use:     "delete",
		Short:   "Delete an existing subscription",
		Example: "  " + env.Cli.Name + " subscription delete my-subscription --namespace joseph-ns",
		Args: ArgValidationConjunction(
			ExactArgs(subscriptionDeleteNumberOfArgs),
			AtPosition(subscriptionDeleteNameIndex, ValidName()),
		),
		RunE: func(cmd *Command, args []string) error {
			deleteOptions.Name = args[subscriptionDeleteNameIndex]
			if err := (*client).DeleteSubscription(deleteOptions); err != nil {
				return err
			}
			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "SUBSCRIPTION_NAME")
	flags := command.Flags()
	flags.StringVarP(&deleteOptions.Namespace, "namespace", "n", "", "the namespace of the subscription")

	return command
}

func SubscriptionList(client *core.Client) *Command {
	listOptions := core.ListSubscriptionsOptions{}

	displayFormat := ""

	command := &Command{
		Use:   "list",
		Short: "List existing subscriptions",
		Example: `  ` + env.Cli.Name + ` subscription list
  ` + env.Cli.Name + ` subscription list --namespace joseph-ns`,
		Args: ExactArgs(subscriptionListNumberOfArgs),
		RunE: func(cmd *Command, args []string) error {
			subscriptions, err := (*client).ListSubscriptions(listOptions)
			if err != nil {
				return err
			}

			if displayFormat == "" {
				Display(cmd.OutOrStdout(), subscriptionToInterfaceSlice(subscriptions.Items), makeSubscriptionExtractors())
				PrintSuccessfulCompletion(cmd)
				return nil
			} else if displayFormat == "dot" {
				return displayAsDot(cmd.OutOrStdout(), subscriptions.Items)
			} else {
				return fmt.Errorf("unsupported output format %q", displayFormat)
			}
		},
	}

	flags := command.Flags()
	flags.StringVarP(&listOptions.Namespace, "namespace", "n", "", "the namespace of the subscriptions")
	flags.StringVarP(&displayFormat, "output", "o", "", "the custom output format to use. Use 'dot' to output graphviz representation")

	return command
}

func displayAsDot(out io.Writer, subscriptions []v1alpha1.Subscription) error {
	tmpl := template.New("dot")
	tmpl.Funcs(map[string]interface{}{"chop": chop})
	tmpl, err := tmpl.Parse(`digraph finite_state_machine {
	rankdir=LR;
	size="8,5"
    node [shape = "box"] {{range .}}{{.Spec.Subscriber}}; {{end}}
    node [shape = "diamond", style = "rounded"] {{range .}}{{.Spec.Channel}}; {{end}}
{{range . -}}
    "{{.Spec.Channel}}" -> "{{.Spec.Subscriber}}";
    {{if ne .Spec.ReplyTo ""}}"{{.Spec.Subscriber}}" -> "{{chop .Spec.ReplyTo }}";{{end}}
{{- end}}
}`)
	if err != nil {
		return err
	}
	return tmpl.Execute(out, subscriptions)
}

// chop removes the '-channel' suffix from the reply-to
func chop(channelName string) (string, error) {
	if strings.HasSuffix(channelName, "-channel") {
		return channelName[0 : len(channelName)-len("-channel")], nil
	} else {
		return "", fmt.Errorf("%q does not end with %q", channelName, "-channel")
	}
}

func defineFlagsForCreate(command *Command, options *core.CreateSubscriptionOptions) {
	flags := command.Flags()
	flags.StringVarP(&options.Subscriber, "subscriber", "s", "", "the subscriber of the subscription")
	flags.StringVarP(&options.Channel, "channel", "c", "", "the input channel of the subscription")
	flags.StringVarP(&options.Reply, "reply", "r", "", "the optional output channel of the subscription")
	flags.StringVarP(&options.Namespace, "namespace", "n", "", "the namespace of the subscription")
	command.MarkFlagRequired("subscriber")
	command.MarkFlagRequired("channel")
}

func computeSubscriptionName(args []string, options core.CreateSubscriptionOptions) string {
	if len(args) == subscriptionCreateMaxNumberOfArgs {
		return args[subscriptionCreateNameIndex]
	}
	return options.Subscriber
}

func subscriptionToInterfaceSlice(subscriptions []v1alpha1.Subscription) []interface{} {
	result := make([]interface{}, len(subscriptions))
	for i := range subscriptions {
		result[i] = subscriptions[i]
	}
	return result
}

func makeSubscriptionExtractors() []NamedExtractor {
	return []NamedExtractor{
		{
			name: "NAME",
			fn:   func(s interface{}) string { return s.(v1alpha1.Subscription).Name },
		},
		{
			name: "CHANNEL",
			fn:   func(s interface{}) string { return s.(v1alpha1.Subscription).Spec.Channel.Name },
		},
		{
			name: "SUBSCRIBER",
			fn: func(s interface{}) string {
				ss := s.(v1alpha1.Subscription).Spec.Subscriber
				if ss == nil {
					return ""
				}
				if ss.Ref != nil {
					return ss.Ref.Name
				}
				if ss.DNSName != nil {
					return *ss.DNSName
				}
				return "<invalid>"
			},
		},
		{
			name: "REPLY",
			fn: func(s interface{}) string {
				r := s.(v1alpha1.Subscription).Spec.Reply
				if r == nil {
					return ""
				}
				return r.Channel.Name
			},
		},
	}
}
