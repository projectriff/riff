package commands

import (
	"fmt"
	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	"github.com/projectriff/riff/pkg/core"
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
		Example: `  riff subscription create --channel tweets --subscriber tweets-logger
  riff subscription create my-subscription --channel tweets --subscriber tweets-logger
  riff subscription create --channel tweets --subscriber tweets-logger --reply-to logged-tweets`,
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
		Example: "  riff subscription delete my-subscription --namespace joseph-ns",
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

	command := &Command{
		Use:   "list",
		Short: "List existing subscriptions",
		Example: `  riff subscription list
  riff subscription list --namespace joseph-ns`,
		Args: ExactArgs(subscriptionListNumberOfArgs),
		RunE: func(cmd *Command, args []string) error {
			subscriptions, err := (*client).ListSubscriptions(listOptions)
			if err != nil {
				return err
			}

			displayList(cmd, subscriptions)
			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}


	flags := command.Flags()
	flags.StringVarP(&listOptions.Namespace, "namespace", "n", "", "the namespace of the subscriptions")

	return command
}

func defineFlagsForCreate(command *Command, options *core.CreateSubscriptionOptions) {
	flags := command.Flags()
	flags.StringVarP(&options.Subscriber, "subscriber", "s", "", "the subscriber of the subscription")
	flags.StringVarP(&options.Channel, "channel", "c", "", "the input channel of the subscription")
	flags.StringVarP(&options.ReplyTo, "reply-to", "r", "", "the optional output channel of the subscription")
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

func displayList(cmd *Command, subscriptions *v1alpha1.SubscriptionList) {
	out := cmd.OutOrStdout()
	display(out, &subscriptions.Items)
	fmt.Fprintln(out)
}

