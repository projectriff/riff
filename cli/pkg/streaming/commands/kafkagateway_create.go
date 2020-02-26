/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/cli/options"
	"github.com/projectriff/cli/pkg/k8s"
	"github.com/projectriff/cli/pkg/race"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KafkaGatewayCreateOptions struct {
	options.ResourceOptions

	BootstrapServers string

	DryRun bool

	Tail        bool
	WaitTimeout time.Duration
}

var (
	_ cli.Validatable = (*KafkaGatewayCreateOptions)(nil)
	_ cli.Executable  = (*KafkaGatewayCreateOptions)(nil)
	_ cli.DryRunable  = (*KafkaGatewayCreateOptions)(nil)
)

func (opts *KafkaGatewayCreateOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	if opts.BootstrapServers == "" {
		errs = errs.Also(cli.ErrMissingField(cli.BootstrapServersFlagName))
	}
	if opts.DryRun && opts.Tail {
		errs = errs.Also(cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName))
	}
	if opts.WaitTimeout < 0 {
		errs = errs.Also(cli.ErrInvalidValue(opts.WaitTimeout, cli.WaitTimeoutFlagName))
	}

	return errs
}

func (opts *KafkaGatewayCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	gateway := &streamv1alpha1.KafkaGateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: streamv1alpha1.KafkaGatewaySpec{
			BootstrapServers: opts.BootstrapServers,
		},
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, gateway, gateway.GetGroupVersionKind())
	} else {
		var err error
		gateway, err = c.StreamingRuntime().KafkaGateways(opts.Namespace).Create(gateway)
		if err != nil {
			return err
		}
	}
	c.Successf("Created kafka gateway %q\n", gateway.Name)
	if opts.Tail {
		c.Infof("Waiting for kafka gateway %q to become ready...\n", gateway.Name)
		err := race.Run(ctx, opts.WaitTimeout,
			func(ctx context.Context) error {
				return k8s.WaitUntilReady(ctx, c.StreamingRuntime().RESTClient(), "kafkagateways", gateway)
			},
			func(ctx context.Context) error {
				return c.Kail.KafkaGatewayLogs(ctx, gateway, cli.TailSinceCreateDefault, c.Stdout)
			},
		)
		if err == context.DeadlineExceeded {
			c.Errorf("Timeout after %q waiting for %q to become ready\n", opts.WaitTimeout, opts.Name)
			c.Infof("To view status run: %s streaming kafka-gateway list %s %s\n", c.Name, cli.NamespaceFlagName, opts.Namespace)
			err = cli.SilenceError(err)
		}
		if err != nil {
			return err
		}
		c.Successf("KafkaGateway %q is ready\n", gateway.Name)
	}

	return nil
}

func (opts *KafkaGatewayCreateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewKafkaGatewayCreateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &KafkaGatewayCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a kafka gateway of messages",
		Long: strings.TrimSpace(`
Creates a Kafka gateway within a namespace.

The gateway is configured with the address of the Kafka broker.
`),
		Example: fmt.Sprintf("%s streaming kafka-gateway create my-kafka-gateway %s kafka.local:9092", c.Name, cli.BootstrapServersFlagName),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.BootstrapServers, cli.StripDash(cli.BootstrapServersFlagName), "", "`address` of the kafka broker")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")
	cmd.Flags().BoolVar(&opts.Tail, cli.StripDash(cli.TailFlagName), false, "watch creation progress")
	cmd.Flags().DurationVar(&opts.WaitTimeout, cli.StripDash(cli.WaitTimeoutFlagName), time.Minute*1, "`duration` to wait for the gateway to become ready when watching progress")

	return cmd
}
