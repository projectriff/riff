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

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/cli/options"
	"github.com/projectriff/riff/cli/pkg/cli/printers"
	knativev1alpha1 "github.com/projectriff/riff/system/pkg/apis/knative/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type DeployerListOptions struct {
	options.ListOptions
}

var (
	_ cli.Validatable = (*DeployerListOptions)(nil)
	_ cli.Executable  = (*DeployerListOptions)(nil)
)

func (opts *DeployerListOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *DeployerListOptions) Exec(ctx context.Context, c *cli.Config) error {
	deployers, err := c.KnativeRuntime().Deployers(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(deployers.Items) == 0 {
		c.Infof("No deployers found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	deployers = deployers.DeepCopy()
	cli.SortByNamespaceAndName(deployers.Items)

	return tablePrinter.PrintObj(deployers, c.Stdout)
}

func NewDeployerListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &DeployerListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of deployers",
		Long: strings.TrimSpace(`
List deployers in a namespace or across all namespaces.

For detail regarding the status of a single deployer, run:

    ` + c.Name + ` knative deployer status <deployer-name>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s knative deployer list", c.Name),
			fmt.Sprintf("%s knative deployer list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func (opts *DeployerListOptions) printList(deployers *knativev1alpha1.DeployerList, printOpts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(deployers.Items))
	for i := range deployers.Items {
		r, err := opts.print(&deployers.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *DeployerListOptions) print(deployer *knativev1alpha1.Deployer, _ printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: deployer},
	}
	refType, refValue := opts.formatRef(deployer)
	var url string
	switch deployer.Spec.IngressPolicy {
	case knativev1alpha1.IngressPolicyClusterLocal:
		if deployer.Status.Address != nil {
			url = deployer.Status.Address.URL
		}
	case knativev1alpha1.IngressPolicyExternal:
		url = deployer.Status.URL
	default:
		url = deployer.Status.URL
	}
	row.Cells = append(row.Cells,
		deployer.Name,
		refType,
		refValue,
		cli.FormatEmptyString(url),
		cli.FormatConditionStatus(deployer.Status.GetCondition(knativev1alpha1.DeployerConditionReady)),
		cli.FormatTimestampSince(deployer.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *DeployerListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Type", Type: "string"},
		{Name: "Ref", Type: "string"},
		{Name: "URL", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}

func (opts *DeployerListOptions) formatRef(deployer *knativev1alpha1.Deployer) (string, string) {
	if deployer.Spec.Build != nil {
		if deployer.Spec.Build.ApplicationRef != "" {
			return "application", deployer.Spec.Build.ApplicationRef
		}
		if deployer.Spec.Build.FunctionRef != "" {
			return "function", deployer.Spec.Build.FunctionRef
		}
		if deployer.Spec.Build.ContainerRef != "" {
			return "container", deployer.Spec.Build.ContainerRef
		}
	} else if deployer.Spec.Template != nil && deployer.Spec.Template.Spec.Containers[0].Image != "" {
		return "image", deployer.Spec.Template.Spec.Containers[0].Image
	}
	return cli.Swarnf("<unknown>"), cli.Swarnf("<unknown>")
}
