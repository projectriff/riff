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
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ApplicationListOptions struct {
	options.ListOptions
}

var (
	_ cli.Validatable = (*ApplicationListOptions)(nil)
	_ cli.Executable  = (*ApplicationListOptions)(nil)
)

func (opts *ApplicationListOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *ApplicationListOptions) Exec(ctx context.Context, c *cli.Config) error {
	applications, err := c.Build().Applications(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(applications.Items) == 0 {
		c.Infof("No applications found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	applications = applications.DeepCopy()
	cli.SortByNamespaceAndName(applications.Items)

	return tablePrinter.PrintObj(applications, c.Stdout)
}

func NewApplicationListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ApplicationListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of applications",
		Long: strings.TrimSpace(`
List applications in a namespace or across all namespaces.

For detail regarding the status of a single application, run:

    ` + c.Name + ` application status <application-name>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s application list", c.Name),
			fmt.Sprintf("%s application list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func (opts *ApplicationListOptions) printList(applications *buildv1alpha1.ApplicationList, printOpts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(applications.Items))
	for i := range applications.Items {
		r, err := opts.print(&applications.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *ApplicationListOptions) print(application *buildv1alpha1.Application, _ printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: application},
	}
	row.Cells = append(row.Cells,
		application.Name,
		cli.FormatEmptyString(application.Status.LatestImage),
		cli.FormatConditionStatus(application.Status.GetCondition(buildv1alpha1.ApplicationConditionReady)),
		cli.FormatTimestampSince(application.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *ApplicationListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Latest Image", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
