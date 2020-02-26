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

type PulsarGatewayCreateOptions struct {
	options.ResourceOptions

	ServiceURL string

	DryRun bool

	Tail        bool
	WaitTimeout time.Duration
}

var (
	_ cli.Validatable = (*PulsarGatewayCreateOptions)(nil)
	_ cli.Executable  = (*PulsarGatewayCreateOptions)(nil)
	_ cli.DryRunable  = (*PulsarGatewayCreateOptions)(nil)
)

func (opts *PulsarGatewayCreateOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	if opts.ServiceURL == "" {
		errs = errs.Also(cli.ErrMissingField(cli.ServiceURLFlagName))
	}
	if opts.DryRun && opts.Tail {
		errs = errs.Also(cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName))
	}
	if opts.WaitTimeout < 0 {
		errs = errs.Also(cli.ErrInvalidValue(opts.WaitTimeout, cli.WaitTimeoutFlagName))
	}

	return errs
}

func (opts *PulsarGatewayCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	gateway := &streamv1alpha1.PulsarGateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: streamv1alpha1.PulsarGatewaySpec{
			ServiceURL: opts.ServiceURL,
		},
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, gateway, gateway.GetGroupVersionKind())
	} else {
		var err error
		gateway, err = c.StreamingRuntime().PulsarGateways(opts.Namespace).Create(gateway)
		if err != nil {
			return err
		}
	}
	c.Successf("Created pulsar gateway %q\n", gateway.Name)
	if opts.Tail {
		c.Infof("Waiting for pulsar gateway %q to become ready...\n", gateway.Name)
		err := race.Run(ctx, opts.WaitTimeout,
			func(ctx context.Context) error {
				return k8s.WaitUntilReady(ctx, c.StreamingRuntime().RESTClient(), "pulsargateways", gateway)
			},
			func(ctx context.Context) error {
				return c.Kail.PulsarGatewayLogs(ctx, gateway, cli.TailSinceCreateDefault, c.Stdout)
			},
		)
		if err == context.DeadlineExceeded {
			c.Errorf("Timeout after %q waiting for %q to become ready\n", opts.WaitTimeout, opts.Name)
			c.Infof("To view status run: %s streaming pulsar-gateway list %s %s\n", c.Name, cli.NamespaceFlagName, opts.Namespace)
			err = cli.SilenceError(err)
		}
		if err != nil {
			return err
		}
		c.Successf("PulsarGateway %q is ready\n", gateway.Name)
	}

	return nil
}

func (opts *PulsarGatewayCreateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewPulsarGatewayCreateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &PulsarGatewayCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a pulsar gateway of messages",
		Long: strings.TrimSpace(`
Creates a Pulsar gateway within a namespace.

The gateway is configured with a Pulsar service URL.
`),
		Example: fmt.Sprintf("%s streaming pulsar-gateway create my-pulsar-gateway %s pulsar://localhost:6650", c.Name, cli.ServiceURLFlagName),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.ServiceURL, cli.StripDash(cli.ServiceURLFlagName), "", "`url` of the pulsar service")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")
	cmd.Flags().BoolVar(&opts.Tail, cli.StripDash(cli.TailFlagName), false, "watch creation progress")
	cmd.Flags().DurationVar(&opts.WaitTimeout, cli.StripDash(cli.WaitTimeoutFlagName), time.Minute*1, "`duration` to wait for the gateway to become ready when watching progress")

	return cmd
}
