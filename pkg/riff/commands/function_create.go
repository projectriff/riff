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

	"github.com/buildpack/pack"
	"github.com/projectriff/riff/pkg/cli"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FunctionCreateOptions struct {
	cli.ResourceOptions

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

func (opts *FunctionCreateOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	if opts.Image == "" {
		errs = errs.Also(cli.ErrMissingField(cli.ImageFlagName))
	}

	if opts.CacheSize != "" {
		// must parse as a resource quantity
		if _, err := resource.ParseQuantity(opts.CacheSize); err != nil {
			errs = errs.Also(cli.ErrInvalidValue(opts.CacheSize, cli.CacheSizeFlagName))
		}
	}

	// git-repo and local-path are mutually exclusive
	if opts.GitRepo == "" && opts.LocalPath == "" {
		errs = errs.Also(cli.ErrMissingOneOf(cli.GitRepoFlagName, cli.LocalPathFlagName))
	} else if opts.GitRepo != "" && opts.LocalPath != "" {
		errs = errs.Also(cli.ErrMultipleOneOf(cli.GitRepoFlagName, cli.LocalPathFlagName))
	}

	// git-revision is required for git-repo
	if opts.GitRepo != "" && opts.GitRevision == "" {
		errs = errs.Also(cli.ErrMissingField(cli.GitRevisionFlagName))
	}

	if opts.LocalPath != "" {
		if opts.SubPath != "" {
			// sub-path cannot be used with local-path
			errs = errs.Also(cli.ErrDisallowedFields(cli.SubPathFlagName))
		}
		if opts.CacheSize != "" {
			// cache-size cannot be used with local-path
			errs = errs.Also(cli.ErrDisallowedFields(cli.CacheSizeFlagName))
		}
	}

	// nothing to do for artifact, handler, and invoker

	return errs
}

func (opts *FunctionCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
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
		function.Spec.Source = &buildv1alpha1.Source{
			Git: &buildv1alpha1.GitSource{
				URL:      opts.GitRepo,
				Revision: opts.GitRevision,
			},
			SubPath: opts.SubPath,
		}
	}
	if opts.LocalPath != "" {
		err := c.Pack.Build(ctx, pack.BuildOptions{
			Image:  opts.Image,
			AppDir: opts.LocalPath,
			// TODO lookup builder from ClusterBuildTemplate/riff-cnb
			Builder: "projectriff/builder:0.2.0",
			Env: map[string]string{
				"RIFF":          "true",
				"RIFF_ARTIFACT": opts.Artifact,
				"RIFF_HANDLER":  opts.Handler,
				"RIFF_OVERRIDE": opts.Invoker,
			},
			Publish: true,
		})
		if err != nil {
			return err
		}
	}

	function, err := c.Build().Functions(opts.Namespace).Create(function)
	if err != nil {
		return err
	}
	c.Successf("Created function %q\n", function.Name)
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
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Image, cli.StripDash(cli.ImageFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.CacheSize, cli.StripDash(cli.CacheSizeFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.Artifact, cli.StripDash(cli.ArtifactFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.Handler, cli.StripDash(cli.HandlerFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.Invoker, cli.StripDash(cli.InvokerFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.LocalPath, cli.StripDash(cli.LocalPathFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.GitRepo, cli.StripDash(cli.GitRepoFlagName), "", "<todo>")
	cmd.Flags().StringVar(&opts.GitRevision, cli.StripDash(cli.GitRevisionFlagName), "master", "<todo>")
	cmd.Flags().StringVar(&opts.SubPath, cli.StripDash(cli.SubPathFlagName), "", "<todo>")

	return cmd
}
