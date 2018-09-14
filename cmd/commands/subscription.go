package commands

import (
	"github.com/projectriff/riff/pkg/core"
	. "github.com/spf13/cobra"
)

const (
	subscriptionCreateNameIndex = iota
	subscriptionCreateMaxNumberOfArgs
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
		Long: `Create a new, optionally named subscription, binding a service to an input channel.
The default name of the subscription is the provided service name.
The service can optionally be bound to an output channel.`,
		Example: `  riff subscription create --from tweets --processor tweets-logger
  riff subscription create my-subscription --from tweets --processor tweets-logger
  riff subscription create --from tweets --processor tweets-logger --to logged-tweets`,
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
			printSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "SUBSCRIPTION_NAME")
	defineFlags(command, &options)
	return command
}

func defineFlags(command *Command, options *core.CreateSubscriptionOptions) {
	flags := command.Flags()
	flags.StringVarP(&options.Subscriber, "processor", "s", "", "the subscriber registered in the subscription")
	flags.StringVarP(&options.Channel, "from", "i", "", "the input channel the service binds to")
	flags.StringVarP(&options.ReplyTo, "to", "o", "", "the optional output channel the service binds to")
	flags.StringVarP(&options.Namespace, "namespace", "n", "", "the namespace of the subscription")
	command.MarkFlagRequired("processor")
	command.MarkFlagRequired("from")
}

func computeSubscriptionName(args []string, options core.CreateSubscriptionOptions) string {
	if len(args) == subscriptionCreateMaxNumberOfArgs {
		return args[subscriptionCreateNameIndex]
	}
	return options.Subscriber
}
