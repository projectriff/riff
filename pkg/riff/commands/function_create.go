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

type FunctionCreateOptions struct {
	Namespace string
	Name      string

	Image    string
	Artifact string
	Handler  string
	Invoker  string

	LocalPath   string
	GitRepo     string
	GitRevision string
	SubPath     string
}

func (opts *FunctionCreateOptions) Validate(ctx context.Context) *apis.FieldError {
	// TODO implement
	return nil
}

func NewFunctionCreateCommand(c *cli.Config) *cobra.Command {
	opts := &FunctionCreateOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "<todo>",
		Example: "<todo>",
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Image, "image", "", "<todo>")
	cmd.Flags().StringVar(&opts.Artifact, "artifact", "", "<todo>")
	cmd.Flags().StringVar(&opts.Handler, "handler", "", "<todo>")
	cmd.Flags().StringVar(&opts.Invoker, "invoker", "", "<todo>")
	cmd.Flags().StringVar(&opts.LocalPath, "local-path", "", "<todo>")
	cmd.Flags().StringVar(&opts.LocalPath, "git-repo", "", "<todo>")
	cmd.Flags().StringVar(&opts.LocalPath, "git-revision", "", "<todo>")
	cmd.Flags().StringVar(&opts.LocalPath, "sub-path", "", "<todo>")

	return cmd
}
