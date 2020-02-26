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
	knativev1alpha1 "github.com/projectriff/system/pkg/apis/knative/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AdapterCreateOptions struct {
	options.ResourceOptions

	ApplicationRef string
	ContainerRef   string
	FunctionRef    string

	ConfigurationRef string
	ServiceRef       string

	Tail        bool
	WaitTimeout string

	DryRun bool
}

var (
	_ cli.Validatable = (*AdapterCreateOptions)(nil)
	_ cli.Executable  = (*AdapterCreateOptions)(nil)
	_ cli.DryRunable  = (*AdapterCreateOptions)(nil)
)

func (opts *AdapterCreateOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	// application-ref, build-ref and container-ref are mutually exclusive
	used := []string{}
	unused := []string{}

	if opts.ApplicationRef != "" {
		used = append(used, cli.ApplicationRefFlagName)
	} else {
		unused = append(unused, cli.ApplicationRefFlagName)
	}

	if opts.ContainerRef != "" {
		used = append(used, cli.ContainerRefFlagName)
	} else {
		unused = append(unused, cli.ContainerRefFlagName)
	}

	if opts.FunctionRef != "" {
		used = append(used, cli.FunctionRefFlagName)
	} else {
		unused = append(unused, cli.FunctionRefFlagName)
	}

	if len(used) == 0 {
		errs = errs.Also(cli.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(cli.ErrMultipleOneOf(used...))
	}

	// configuration-ref and service-ref are mutually exclusive
	used = []string{}
	unused = []string{}

	if opts.ConfigurationRef != "" {
		used = append(used, cli.ConfigurationRefFlagName)
	} else {
		unused = append(unused, cli.ConfigurationRefFlagName)
	}

	if opts.ServiceRef != "" {
		used = append(used, cli.ServiceRefFlagName)
	} else {
		unused = append(unused, cli.ServiceRefFlagName)
	}

	if len(used) == 0 {
		errs = errs.Also(cli.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(cli.ErrMultipleOneOf(used...))
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

	return errs
}

func (opts *AdapterCreateOptions) Exec(ctx context.Context, c *cli.Config) error {
	adapter := &knativev1alpha1.Adapter{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
		Spec: knativev1alpha1.AdapterSpec{},
	}

	if opts.ApplicationRef != "" {
		adapter.Spec.Build = knativev1alpha1.Build{
			ApplicationRef: opts.ApplicationRef,
		}
	}
	if opts.ContainerRef != "" {
		adapter.Spec.Build = knativev1alpha1.Build{
			ContainerRef: opts.ContainerRef,
		}
	}
	if opts.FunctionRef != "" {
		adapter.Spec.Build = knativev1alpha1.Build{
			FunctionRef: opts.FunctionRef,
		}
	}

	if opts.ConfigurationRef != "" {
		adapter.Spec.Target = knativev1alpha1.AdapterTarget{
			ConfigurationRef: opts.ConfigurationRef,
		}
	}
	if opts.ServiceRef != "" {
		adapter.Spec.Target = knativev1alpha1.AdapterTarget{
			ServiceRef: opts.ServiceRef,
		}
	}

	if opts.DryRun {
		cli.DryRunResource(ctx, adapter, adapter.GetGroupVersionKind())
	} else {
		var err error
		adapter, err = c.KnativeRuntime().Adapters(opts.Namespace).Create(adapter)
		if err != nil {
			return err
		}
	}
	c.Successf("Created adapter %q\n", adapter.Name)
	if opts.Tail {
		c.Infof("Waiting for adapter %q to become ready...\n", adapter.Name)
		// err guarded by Validate()
		timeout, _ := time.ParseDuration(opts.WaitTimeout)
		err := race.Run(ctx, timeout,
			func(ctx context.Context) error {
				return k8s.WaitUntilReady(ctx, c.KnativeRuntime().RESTClient(), "adapters", adapter)
			},
		)
		if err == context.DeadlineExceeded {
			c.Errorf("Timeout after %q waiting for %q to become ready\n", opts.WaitTimeout, opts.Name)
			c.Infof("To view status run: %s knative adapter list %s %s\n", c.Name, cli.NamespaceFlagName, opts.Namespace)
			err = cli.SilenceError(err)
		}
		if err != nil {
			return err
		}
		c.Successf("Adapter %q is ready\n", adapter.Name)
	}
	return nil
}

func (opts *AdapterCreateOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewAdapterCreateCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &AdapterCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create an adapter to Knative Serving",
		Long: strings.TrimSpace(`
Create a new adapter by watching a build for the latest image, pushing those
images to a target Knative Service or Configuration.

No new Knative resources are created directly by the adapter, it only updates
the image for an existing resource.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s knative adapter create my-adapter %s my-app %s my-kservice", c.Name, cli.ApplicationRefFlagName, cli.ServiceRefFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.Args(cmd,
		cli.NameArg(&opts.Name),
	)

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.ApplicationRef, cli.StripDash(cli.ApplicationRefFlagName), "", "`name` of application to deploy")
	_ = cmd.MarkFlagCustom(cli.StripDash(cli.ApplicationRefFlagName), "__"+c.Name+"_list_applications")
	cmd.Flags().StringVar(&opts.ContainerRef, cli.StripDash(cli.ContainerRefFlagName), "", "`name` of container to deploy")
	_ = cmd.MarkFlagCustom(cli.StripDash(cli.ContainerRefFlagName), "__"+c.Name+"_list_containers")
	cmd.Flags().StringVar(&opts.FunctionRef, cli.StripDash(cli.FunctionRefFlagName), "", "`name` of function to deploy")
	_ = cmd.MarkFlagCustom(cli.StripDash(cli.FunctionRefFlagName), "__"+c.Name+"_list_functions")
	cmd.Flags().StringVar(&opts.ConfigurationRef, cli.StripDash(cli.ConfigurationRefFlagName), "", "`name` of Knative configuration to update")
	_ = cmd.MarkFlagCustom(cli.StripDash(cli.ConfigurationRefFlagName), "__"+c.Name+"_list_knative_configurations")
	cmd.Flags().StringVar(&opts.ServiceRef, cli.StripDash(cli.ServiceRefFlagName), "", "`name` of Knative service to update")
	_ = cmd.MarkFlagCustom(cli.StripDash(cli.ServiceRefFlagName), "__"+c.Name+"_list_knative_services")
	cmd.Flags().BoolVar(&opts.Tail, cli.StripDash(cli.TailFlagName), false, "watch adapter logs")
	cmd.Flags().StringVar(&opts.WaitTimeout, cli.StripDash(cli.WaitTimeoutFlagName), "10m", "`duration` to wait for the adapter to become ready when watching logs")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")

	return cmd
}
