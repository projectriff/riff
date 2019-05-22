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

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/cli/printers"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type RouteListOptions struct {
	cli.ListOptions
}

func (opts *RouteListOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *RouteListOptions) Exec(ctx context.Context, c *cli.Config) error {
	routes, err := c.Request().Routes(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(routes.Items) == 0 {
		c.Infof("No routes found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := printRouteColumns()
		h.TableHandler(columns, printRouteList)
		h.TableHandler(columns, printRoute)
	})

	routes = routes.DeepCopy()
	cli.SortByNamespaceAndName(routes.Items)

	return tablePrinter.PrintObj(routes, c.Stdout)
}

func NewRouteListCommand(c *cli.Config) *cobra.Command {
	opts := &RouteListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "<todo>",
		Example: "<todo>",
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func printRouteList(routes *requestv1alpha1.RouteList, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(routes.Items))
	for i := range routes.Items {
		rows = append(rows, printRoute(&routes.Items[i], opts)...)
	}
	return rows, nil
}

func printRoute(route *requestv1alpha1.Route, opts printers.PrintOptions) []metav1beta1.TableRow {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: route},
	}
	refType, refValue := routeRef(route)
	row.Cells = append(row.Cells,
		route.Name,
		refType,
		refValue,
		cli.FormatEmptyString(route.Status.Domain),
		cli.FormatConditionStatus(route.Status.GetCondition(requestv1alpha1.RouteConditionReady)),
		cli.FormatTimestampSince(route.CreationTimestamp),
	)
	return []metav1beta1.TableRow{row}
}

func printRouteColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Type", Type: "string"},
		{Name: "Ref", Type: "string"},
		{Name: "Domain", Type: "string"},
		{Name: "Ready", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}

func routeRef(route *requestv1alpha1.Route) (string, string) {
	if route.Spec.Build != nil {
		if route.Spec.Build.ApplicationRef != "" {
			return "application", route.Spec.Build.ApplicationRef
		}
		if route.Spec.Build.FunctionRef != "" {
			return "function", route.Spec.Build.FunctionRef
		}
	} else if route.Spec.Template != nil && route.Spec.Template.Containers[0].Image != "" {
		return "image", route.Spec.Template.Containers[0].Image
	}
	return cli.Swarnf("<unknown>"), cli.Swarnf("<unknown>")
}
