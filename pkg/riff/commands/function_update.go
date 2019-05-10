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
	"k8s.io/apimachinery/pkg/api/validation"
)

type FunctionUpdateOptions struct {
	Namespace string
	Name      string

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

	if opts.Namespace == "" {
		errs = errs.Also(apis.ErrMissingField("namespace"))
	}

	if opts.Name == "" {
		errs = errs.Also(apis.ErrMissingField("name"))
	} else {
		if out := validation.NameIsDNSSubdomain(opts.Name, false); len(out) != 0 {
			// TODO capture info about why the name is invalid
			errs = errs.Also(apis.ErrInvalidValue(opts.Name, "name"))
		}
	}

	// TODO validate other fields

	return errs
}

func NewFunctionUpdateCommand(c *cli.Config) *cobra.Command {
	opts := &FunctionUpdateOptions{}

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "<todo>",
		Example: "<todo>",
		Args:    cli.Args(),
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
	cmd.Flags().StringVar(&opts.GitRepo, "git-repo", "", "<todo>")
	cmd.Flags().StringVar(&opts.GitRevision, "git-revision", "", "<todo>")
	cmd.Flags().StringVar(&opts.SubPath, "sub-path", "", "<todo>")

	return cmd
}
