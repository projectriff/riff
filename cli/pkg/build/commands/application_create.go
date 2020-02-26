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
	"runtime"
	"strings"
	"time"

	"github.com/buildpacks/pack"
	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/cli/options"
	"github.com/projectriff/cli/pkg/k8s"
	"github.com/projectriff/cli/pkg/parsers"
	"github.com/projectriff/cli/pkg/race"
	"github.com/projectriff/cli/pkg/validation"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplicationCreateOptions struct {
	options.ResourceOptions

	Image     string
	CacheSize string

	LocalPath   string
	GitRepo     string
	GitRevision string
	SubPath     string

	Env []string

	LimitCPU    string
	LimitMemory string

	Tail        bool
	WaitTimeout string

	DryRun bool
}

var (
	_ cli.Validatable = (*ApplicationCreateOptions)(nil)
	_ cli.Executable  = (*ApplicationCreateOptions)(nil)
	_ cli.DryRunable  = (*ApplicationCreateOptions)(nil)
)

func (opts *ApplicationCreateOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

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
			errs = errs.Also(cli.ErrDisallowedFields(cli.SubPathFlagName, ""))
		}
		if opts.CacheSize != "" {
			// cache-size cannot be used with local-path
			errs = errs.Also(cli.ErrDisallowedFields(cli.CacheSizeFlagName, ""))
		}
	}

	errs = errs.Also(validation.EnvVars(opts.Env, cli.EnvFlagName))

	if opts.LimitCPU != "" {
		errs = errs.Also(validation.Quantity(opts.LimitCPU, cli.LimitCPUFlagName))
	}
	if opts.LimitMemory != "" {
		errs = errs.Also(validation.Quantity(opts.LimitMemory, cli.LimitMemoryFlagName))
	}

	if opts.Tail {
		if opts.WaitTimeout == "" {
			errs = errs.Also(cli.ErrMissingField(cli.WaitTimeoutFlagName))
		} else if _, err := time.ParseDuration(opts.WaitTimeout); err != nil {
			errs = errs.Also(cli.ErrInvalidValue(opts.WaitTimeout, cli.WaitTimeoutFlagName))
		}
	}

	if opts.DryRun && opts.Tail {
		errs = errs.Also(cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName))
	}

	if opts.LocalPath != "" && runtime.GOOS == "windows" {
		errs = errs.Also(cli.ErrInvalidValue(fmt.Sprintf("%s is not available on Windows", cli.LocalPathFlagName), cli.LocalPathFlagName))
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
			Git: &buildv1alpha1.Git{
				URL:      opts.GitRepo,
				Revision: opts.GitRevision,
			},
			SubPath: opts.SubPath,
		}
	}

	for _, env := range opts.Env {
		if application.Spec.Build.Env == nil {
			application.Spec.Build.Env = []corev1.EnvVar{}
		}
		application.Spec.Build.Env = append(application.Spec.Build.Env, parsers.EnvVar(env))
	}

	if (opts.LimitCPU != "" || opts.LimitMemory != "") && application.Spec.Build.Resources.Limits == nil {
		application.Spec.Build.Resources.Limits = corev1.ResourceList{}
	}
	if opts.LimitCPU != "" {
		// parse errors are handled by the opt validation
		application.Spec.Build.Resources.Limits[corev1.ResourceCPU] = resource.MustParse(opts.LimitCPU)
	}
	if opts.LimitMemory != "" {
		// parse errors are handled by the opt validation
		application.Spec.Build.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(opts.LimitMemory)
	}

	if opts.LocalPath != "" {
		targetImage := opts.Image
		if strings.HasPrefix(targetImage, "_") {
			riffBuildConfig, err := c.Core().ConfigMaps(application.Namespace).Get("riff-build", metav1.GetOptions{})
			if err != nil {
				if apierrs.IsNotFound(err) {
					return fmt.Errorf("default image prefix requires initialized credentials, run `%s help credentials`", c.Name)
				}
				return err
			}
			targetImage, err = buildv1alpha1.ResolveDefaultImage(application, riffBuildConfig.Data["default-image-prefix"])
			if err != nil {
				return err
			}
		}
		builders, err := c.Core().ConfigMaps("riff-system").Get("builders", metav1.GetOptions{})
		if err != nil {
			return err
		}
		builder := builders.Data["riff-application"]
		if builder == "" {
			return fmt.Errorf("unknown builder for %q", "riff-application")
		}
		env := map[string]string{}
		for _, envvar := range application.Spec.Build.Env {
			env[envvar.Name] = envvar.Value
		}
		err = c.Pack.Build(ctx, pack.BuildOptions{
			Image:   targetImage,
			AppPath: opts.LocalPath,
			Builder: builder,
			Env:     env,
			Publish: true,
		})
		if err != nil {
			return err
		}
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, application, application.GetGroupVersionKind())
	} else {
		var err error
		application, err = c.Build().Applications(opts.Namespace).Create(application)
		if err != nil {
			return err
		}
	}
	c.Successf("Created application %q\n", application.Name)
	if opts.Tail {
		c.Infof("Waiting for application %q to become ready...\n", application.Name)
		// err guarded by Validate()
		timeout, _ := time.ParseDuration(opts.WaitTimeout)
		err := race.Run(ctx, timeout,
			func(ctx context.Context) error {
				return k8s.WaitUntilReady(ctx, c.Build().RESTClient(), "applications", application)
			},
			func(ctx context.Context) error {
				return c.Kail.ApplicationLogs(ctx, application, cli.TailSinceCreateDefault, c.Stdout)
			},
		)
		if err == context.DeadlineExceeded {
			c.Errorf("Timeout after %q waiting for %q to become ready\n", opts.WaitTimeout, opts.Name)
			c.Infof("To view status run: %s application list %s %s\n", c.Name, cli.NamespaceFlagName, opts.Namespace)
			c.Infof("To continue watching logs run: %s application tail %s %s %s\n", c.Name, opts.Name, cli.NamespaceFlagName, opts.Namespace)
			err = cli.SilenceError(err)
		}
		if err != nil {
			return err
		}
		c.Successf("Application %q is ready\n", application.Name)
	}
	return nil
}

func (opts *ApplicationCreateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewApplicationCreateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ApplicationCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create an application from source",
		Long: strings.TrimSpace(`
Create an application from source using the application Cloud Native Buildpack
builder.

Application source can be specified either as a Git repository or as a local
directory. Builds from Git are run in the cluster while builds from a local
directory are run inside a local Docker daemon and are orchestrated by this
command (in the future, builds from local source may also be run in the
cluster).
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s application create my-app %s registry.example.com/image %s https://example.com/my-app.git", c.Name, cli.ImageFlagName, cli.GitRepoFlagName),
			fmt.Sprintf("%s application create my-app %s registry.example.com/image %s ./my-app", c.Name, cli.ImageFlagName, cli.LocalPathFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.Image, cli.StripDash(cli.ImageFlagName), "_", "`repository` where the built images are pushed")
	cmd.Flags().StringVar(&opts.CacheSize, cli.StripDash(cli.CacheSizeFlagName), "", "`size` of persistent volume to cache resources between builds")
	cmd.Flags().StringVar(&opts.LocalPath, cli.StripDash(cli.LocalPathFlagName), "", "path to `directory` containing source code on the local machine")
	_ = cmd.MarkFlagDirname(cli.StripDash(cli.LocalPathFlagName))
	cmd.Flags().StringVar(&opts.GitRepo, cli.StripDash(cli.GitRepoFlagName), "", "git `url` to remote source code")
	cmd.Flags().StringVar(&opts.GitRevision, cli.StripDash(cli.GitRevisionFlagName), "master", "`refspec` within the git repo to checkout")
	cmd.Flags().StringVar(&opts.SubPath, cli.StripDash(cli.SubPathFlagName), "", "path to `directory` within the git repo to checkout")
	cmd.Flags().StringArrayVar(&opts.Env, cli.StripDash(cli.EnvFlagName), []string{}, fmt.Sprintf("environment `variable` defined as a key value pair separated by an equals sign, example %q (may be set multiple times)", fmt.Sprintf("%s MY_VAR=my-value", cli.EnvFlagName)))
	cmd.Flags().StringVar(&opts.LimitCPU, cli.StripDash(cli.LimitCPUFlagName), "", "the maximum amount of cpu allowed, in CPU `cores` (500m = .5 cores)")
	cmd.Flags().StringVar(&opts.LimitMemory, cli.StripDash(cli.LimitMemoryFlagName), "", "the maximum amount of memory allowed, in `bytes` (500Mi = 500MiB = 500 * 1024 * 1024)")
	cmd.Flags().BoolVar(&opts.Tail, cli.StripDash(cli.TailFlagName), false, "watch build logs")
	cmd.Flags().StringVar(&opts.WaitTimeout, cli.StripDash(cli.WaitTimeoutFlagName), "10m", "`duration` to wait for the application to become ready when watching logs")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")

	return cmd
}
