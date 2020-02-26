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
	"github.com/projectriff/cli/pkg/cli/printers"
	knativev1alpha1 "github.com/projectriff/system/pkg/apis/knative/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AdapterListOptions struct {
	options.ListOptions
}

var (
	_ cli.Validatable = (*AdapterListOptions)(nil)
	_ cli.Executable  = (*AdapterListOptions)(nil)
)

func (opts *AdapterListOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *AdapterListOptions) Exec(ctx context.Context, c *cli.Config) error {
	adapters, err := c.KnativeRuntime().Adapters(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(adapters.Items) == 0 {
		c.Infof("No adapters found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	adapters = adapters.DeepCopy()
	cli.SortByNamespaceAndName(adapters.Items)

	return tablePrinter.PrintObj(adapters, c.Stdout)
}

func NewAdapterListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &AdapterListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of adapters",
		Long: strings.TrimSpace(`
List adapters in a namespace or across all namespaces.

For detail regarding the status of a single adapter, run:

    ` + c.Name + ` knative adapter status <adapter-name>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s knative adapter list", c.Name),
			fmt.Sprintf("%s knative adapter list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func (opts *AdapterListOptions) printList(adapters *knativev1alpha1.AdapterList, printOpts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(adapters.Items))
	for i := range adapters.Items {
		r, err := opts.print(&adapters.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *AdapterListOptions) print(adapter *knativev1alpha1.Adapter, _ printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: adapter},
	}
	buildRefType, buildRefValue := opts.formatBuildRef(adapter)
	targetRefType, targetRefValue := opts.formatTargetRef(adapter)
	row.Cells = append(row.Cells,
		adapter.Name,
		buildRefType,
		buildRefValue,
		targetRefType,
		targetRefValue,
		cli.FormatConditionStatus(adapter.Status.GetCondition(knativev1alpha1.AdapterConditionReady)),
		cli.FormatTimestampSince(adapter.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *AdapterListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Build Type", Type: "string"},
		{Name: "Build Ref", Type: "string"},
		{Name: "Target Type", Type: "string"},
		{Name: "Target Ref", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}

func (opts *AdapterListOptions) formatBuildRef(adapter *knativev1alpha1.Adapter) (string, string) {
	if adapter.Spec.Build.ApplicationRef != "" {
		return "application", adapter.Spec.Build.ApplicationRef
	}
	if adapter.Spec.Build.FunctionRef != "" {
		return "function", adapter.Spec.Build.FunctionRef
	}
	if adapter.Spec.Build.ContainerRef != "" {
		return "container", adapter.Spec.Build.ContainerRef
	}
	return cli.Swarnf("<unknown>"), cli.Swarnf("<unknown>")
}

func (opts *AdapterListOptions) formatTargetRef(adapter *knativev1alpha1.Adapter) (string, string) {
	if adapter.Spec.Target.ConfigurationRef != "" {
		return "configuration", adapter.Spec.Target.ConfigurationRef
	}
	if adapter.Spec.Target.ServiceRef != "" {
		return "service", adapter.Spec.Target.ServiceRef
	}
	return cli.Swarnf("<unknown>"), cli.Swarnf("<unknown>")
}
