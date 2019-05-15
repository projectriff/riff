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

	"github.com/knative/pkg/apis"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
)

type FunctionUpdateOptions struct {
	cli.ResourceOptions

	Image    string
	Artifact string
	Handler  string
	Invoker  string

	GitRepo     string
	GitRevision string
	SubPath     string
}

func (opts *FunctionUpdateOptions) Validate(ctx context.Context) *apis.FieldError {
	errs := &apis.FieldError{}

	errs = errs.Also(opts.ResourceOptions.Validate((ctx)))

	// TODO validate other fields

	return errs
}

func (opts *FunctionUpdateOptions) Exec(ctx context.Context, c *cli.Config) error {
	return fmt.Errorf("not implemented")
}

func NewFunctionUpdateCommand(c *cli.Config) *cobra.Command {
	opts := &FunctionUpdateOptions{}

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "<todo>",
		Example: "<todo>",
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Image, cli.StripDash(cli.ImageFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.Artifact, cli.StripDash(cli.ArtifactFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.Handler, cli.StripDash(cli.HandlerFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.Invoker, cli.StripDash(cli.InvokerFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.GitRepo, cli.StripDash(cli.GitRepoFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.GitRevision, cli.StripDash(cli.GitRevisionFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.SubPath, cli.StripDash(cli.SubPathFlagName), "", "<todo>")

	return cmd
}
