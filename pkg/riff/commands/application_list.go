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

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/cli/printers"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ApplicationListOptions struct {
	cli.ListOptions
}

var (
	_ cli.Validatable = (*ApplicationListOptions)(nil)
	_ cli.Executable  = (*ApplicationListOptions)(nil)
)

func (opts *ApplicationListOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

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
		columns := printApplicationColumns()
		h.TableHandler(columns, printApplicationList)
		h.TableHandler(columns, printApplication)
	})

	applications = applications.DeepCopy()
	cli.SortByNamespaceAndName(applications.Items)

	return tablePrinter.PrintObj(applications, c.Stdout)
}

func NewApplicationListCommand(c *cli.Config) *cobra.Command {
	opts := &ApplicationListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of applications",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s application list", c.Name),
			fmt.Sprintf("%s application list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func printApplicationList(applications *buildv1alpha1.ApplicationList, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(applications.Items))
	for i := range applications.Items {
		r, err := printApplication(&applications.Items[i], opts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printApplication(application *buildv1alpha1.Application, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
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

func printApplicationColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Latest Image", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
