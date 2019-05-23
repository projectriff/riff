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

	"github.com/buildpack/pack"
	"github.com/projectriff/riff/pkg/cli"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplicationCreateOptions struct {
	cli.ResourceOptions

	Image     string
	CacheSize string

	LocalPath   string
	GitRepo     string
	GitRevision string
	SubPath     string
}

func (opts *ApplicationCreateOptions) Validate(ctx context.Context) *cli.FieldError {
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

	return errs
}

func (opts *ApplicationCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	application := &buildv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: buildv1alpha1.ApplicationSpec{
			Image: opts.Image,
		},
	}
	if opts.CacheSize != "" {
		quantity := resource.MustParse(opts.CacheSize)
		application.Spec.CacheSize = &quantity
	}
	if opts.GitRepo != "" {
		application.Spec.Source = &buildv1alpha1.Source{
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
			// TODO lookup builder from ClusterBuildTemplate/cf-cnb
			Builder: "cloudfoundry/cnb:bionic",
			Publish: true,
		})
		if err != nil {
			return err
		}
	}

	application, err := c.Build().Applications(opts.Namespace).Create(application)
	if err != nil {
		return err
	}
	c.Successf("Created application %q\n", application.Name)
	return nil
}

func NewApplicationCreateCommand(c *cli.Config) *cobra.Command {
	opts := &ApplicationCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "build an application from source",
		Long: `
<todo>
`,
		Example: strings.Join([]string{
			fmt.Sprintf("%s application create my-app %s registry.example.com/image %s https://example.com/my-app.git", c.Name, cli.ImageFlagName, cli.GitRepoFlagName),
			fmt.Sprintf("%s application create my-app %s registry.example.com/image %s ./my-app", c.Name, cli.ImageFlagName, cli.LocalPathFlagName),
		}, "\n"),
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Image, cli.StripDash(cli.ImageFlagName), "", "repository where the built images are pushed")
	cmd.Flags().StringVar(&opts.CacheSize, cli.StripDash(cli.CacheSizeFlagName), "", "size of persistent volume to cache resources between builds")
	cmd.Flags().StringVar(&opts.LocalPath, cli.StripDash(cli.LocalPathFlagName), "", "path to source code on the local machine")
	cmd.Flags().StringVar(&opts.GitRepo, cli.StripDash(cli.GitRepoFlagName), "", "git url to remote source code")
	cmd.Flags().StringVar(&opts.GitRevision, cli.StripDash(cli.GitRevisionFlagName), "master", "refspec within the git repo to checkout")
	cmd.Flags().StringVar(&opts.SubPath, cli.StripDash(cli.SubPathFlagName), "", "path within the git repo to checkout")

	return cmd
}
