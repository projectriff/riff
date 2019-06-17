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

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/validation"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StreamCreateOptions struct {
	cli.ResourceOptions

	Provider    string
	ContentType string
}

func (opts *StreamCreateOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	if opts.Provider == "" {
		errs = errs.Also(cli.ErrMissingField(cli.ProviderFlagName))
	}

	contentType := opts.ContentType
	if contentType != "" {
		errs = errs.Also(validation.MimeType(contentType, cli.ContentTypeName))
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
			Provider:    opts.Provider,
			ContentType: opts.ContentType,
		},
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, stream, stream.GetGroupVersionKind())
	} else {
		var err error
		stream, err = c.Stream().Streams(opts.Namespace).Create(stream)
		if err != nil {
			return err
		}
	}
	c.Successf("Created stream %q\n", stream.Name)
	return nil
}

func NewStreamCreateCommand(c *cli.Config) *cobra.Command {
	opts := &StreamCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a stream of messages",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: fmt.Sprintf("%s stream create %s my-provider", c.Name, cli.ProviderFlagName),
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Provider, cli.StripDash(cli.ProviderFlagName), "", "`name` of stream provider")
	cmd.Flags().StringVar(&opts.ContentType, cli.StripDash(cli.ContentTypeName), "", "`MIME type` for message payloads accepted by the stream")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")

	return cmd
}
