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
	"github.com/projectriff/cli/pkg/validation"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StreamCreateOptions struct {
	options.ResourceOptions

	Gateway     string
	ContentType string

	DryRun bool

	Tail        bool
	WaitTimeout time.Duration
}

var (
	_ cli.Validatable = (*StreamCreateOptions)(nil)
	_ cli.Executable  = (*StreamCreateOptions)(nil)
	_ cli.DryRunable  = (*StreamCreateOptions)(nil)
)

func (opts *StreamCreateOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	if opts.Gateway == "" {
		errs = errs.Also(cli.ErrMissingField(cli.GatewayFlagName))
	}

	contentType := opts.ContentType
	if contentType != "" {
		errs = errs.Also(validation.MimeType(contentType, cli.ContentTypeFlagName))
	}

	if opts.DryRun && opts.Tail {
		errs = errs.Also(cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName))
	}
	if opts.WaitTimeout < 0 {
		errs = errs.Also(cli.ErrInvalidValue(opts.WaitTimeout, cli.WaitTimeoutFlagName))
	}
	return errs
}

func (opts *StreamCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	stream := &streamv1alpha1.Stream{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: streamv1alpha1.StreamSpec{
			Gateway:     corev1.LocalObjectReference{Name: opts.Gateway},
			ContentType: opts.ContentType,
		},
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, stream, stream.GetGroupVersionKind())
	} else {
		var err error
		stream, err = c.StreamingRuntime().Streams(opts.Namespace).Create(stream)
		if err != nil {
			return err
		}
	}
	c.Successf("Created stream %q\n", stream.Name)
	if opts.Tail {
		c.Infof("Waiting for stream %q to become ready...\n", stream.Name)
		err := race.Run(ctx, opts.WaitTimeout,
			func(ctx context.Context) error {
				return k8s.WaitUntilReady(ctx, c.StreamingRuntime().RESTClient(), "streams", stream)
			},
		)
		if err == context.DeadlineExceeded {
			c.Errorf("Timeout after %q waiting for %q to become ready\n", opts.WaitTimeout, opts.Name)
			c.Infof("To view status run: %s streaming stream list %s %s\n", c.Name, cli.NamespaceFlagName, opts.Namespace)
			err = cli.SilenceError(err)
		}
		if err != nil {
			return err
		}
		c.Successf("Stream %q is ready\n", stream.Name)
	}
	return nil
}

func (opts *StreamCreateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewStreamCreateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &StreamCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a stream of messages",
		Long: strings.TrimSpace(`
Create a stream resource within a namespace and provision a stream in the
underlying message broker via the referenced stream gateway.

The created stream can then be referenced as an input or an output of a given
function when creating a streaming processor.
`),
		Example: fmt.Sprintf("%s streaming stream create my-stream %s my-gateway", c.Name, cli.GatewayFlagName),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Gateway, cli.StripDash(cli.GatewayFlagName), "", "`name` of stream gateway")
	_ = cmd.MarkFlagCustom(cli.StripDash(cli.GatewayFlagName), "__"+c.Name+"_list_streaming_gateways")
	cmd.Flags().StringVar(&opts.ContentType, cli.StripDash(cli.ContentTypeFlagName), "", "`MIME type` for message payloads accepted by the stream")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")
	cmd.Flags().BoolVar(&opts.Tail, cli.StripDash(cli.TailFlagName), false, "watch provisioning progress")
	cmd.Flags().DurationVar(&opts.WaitTimeout, cli.StripDash(cli.WaitTimeoutFlagName), time.Second*10, "`duration` to wait for the stream to become ready when watching progress")

	return cmd
}
