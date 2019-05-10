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
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FunctionCreateOptions struct {
	Namespace string
	Name      string

	Image     string
	CacheSize string

	Artifact string
	Handler  string
	Invoker  string

	LocalPath   string
	GitRepo     string
	GitRevision string
	SubPath     string
}

func (opts *FunctionCreateOptions) Validate(ctx context.Context) *apis.FieldError {
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

	if opts.Image == "" {
		errs = errs.Also(apis.ErrMissingField("image"))
	} else if false {
		// TODO validate image
	}

	if opts.CacheSize != "" {
		// must parse as a resource quantity
		if _, err := resource.ParseQuantity(opts.CacheSize); err != nil {
			errs = errs.Also(apis.ErrInvalidValue(opts.CacheSize, "cache-size"))
		}
	}

	// git-repo and local-path are mutually exclusive
	if opts.GitRepo == "" && opts.LocalPath == "" {
		errs = errs.Also(apis.ErrMissingOneOf("git-repo", "local-path"))
	} else if opts.GitRepo != "" && opts.LocalPath != "" {
		errs = errs.Also(apis.ErrMultipleOneOf("git-repo", "local-path"))
	}

	// git-revision is required for git-repo
	if opts.GitRepo != "" && opts.GitRevision == "" {
		errs = errs.Also(apis.ErrMissingField("git-revision"))
	}

	// sub-path cannot be used with local-path
	if opts.SubPath != "" && opts.LocalPath != "" {
		errs = errs.Also(apis.ErrDisallowedFields("sub-path"))
	}

	// nothing to do for artifact, handler, and invoker

	return errs
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
			function := &buildv1alpha1.Function{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: opts.Namespace,
					Name:      opts.Name,
				},
				Spec: buildv1alpha1.FunctionSpec{
					Image:    opts.Image,
					Artifact: opts.Artifact,
					Handler:  opts.Handler,
					Invoker:  opts.Invoker,
				},
			}
			if opts.CacheSize != "" {
				quantity := resource.MustParse(opts.CacheSize)
				function.Spec.CacheSize = &quantity
			}
			if opts.GitRepo != "" {
				function.Spec.Source = buildv1alpha1.Source{
					Git: &buildv1alpha1.GitSource{
						URL:      opts.GitRepo,
						Revision: opts.GitRevision,
					},
					SubPath: opts.SubPath,
				}
			}
			if opts.LocalPath != "" {
				// TODO implement
				return fmt.Errorf("not implemented")
			}

			function, err := c.Build().Functions(opts.Namespace).Create(function)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created function %q\n", function.Name)
			return nil
		},
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Image, "image", "", "<todo>")
	cmd.Flags().StringVar(&opts.CacheSize, "cache-size", "", "<todo>")
	cmd.Flags().StringVar(&opts.Artifact, "artifact", "", "<todo>")
	cmd.Flags().StringVar(&opts.Handler, "handler", "", "<todo>")
	cmd.Flags().StringVar(&opts.Invoker, "invoker", "", "<todo>")
	cmd.Flags().StringVar(&opts.LocalPath, "local-path", "", "<todo>")
	cmd.Flags().StringVar(&opts.GitRepo, "git-repo", "", "<todo>")
	cmd.Flags().StringVar(&opts.GitRevision, "git-revision", "master", "<todo>")
	cmd.Flags().StringVar(&opts.SubPath, "sub-path", "", "<todo>")

	return cmd
}
