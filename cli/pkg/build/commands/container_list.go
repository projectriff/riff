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
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ContainerListOptions struct {
	options.ListOptions
}

var (
	_ cli.Validatable = (*ContainerListOptions)(nil)
	_ cli.Executable  = (*ContainerListOptions)(nil)
)

func (opts *ContainerListOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *ContainerListOptions) Exec(ctx context.Context, c *cli.Config) error {
	containers, err := c.Build().Containers(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(containers.Items) == 0 {
		c.Infof("No containers found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	containers = containers.DeepCopy()
	cli.SortByNamespaceAndName(containers.Items)

	return tablePrinter.PrintObj(containers, c.Stdout)
}

func NewContainerListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ContainerListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of containers",
		Long: strings.TrimSpace(`
List containers in a namespace or across all namespaces.

For detail regarding the status of a single container, run:

    ` + c.Name + ` container status <container-name>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s container list", c.Name),
			fmt.Sprintf("%s container list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func (opts *ContainerListOptions) printList(containers *buildv1alpha1.ContainerList, printOpts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(containers.Items))
	for i := range containers.Items {
		r, err := opts.print(&containers.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *ContainerListOptions) print(container *buildv1alpha1.Container, _ printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: container},
	}
	row.Cells = append(row.Cells,
		container.Name,
		cli.FormatEmptyString(container.Status.LatestImage),
		cli.FormatConditionStatus(container.Status.GetCondition(buildv1alpha1.ContainerConditionReady)),
		cli.FormatTimestampSince(container.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *ContainerListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Latest Image", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
