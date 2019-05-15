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

type RequestProcessorListOptions struct {
	cli.ListOptions
}

func (opts *RequestProcessorListOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := &cli.FieldError{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *RequestProcessorListOptions) Exec(ctx context.Context, c *cli.Config) error {
	requestprocessors, err := c.Request().RequestProcessors(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(requestprocessors.Items) == 0 {
		c.Infof("No request processors found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := printRequestProcessorColumns()
		h.TableHandler(columns, printRequestProcessorList)
		h.TableHandler(columns, printRequestProcessor)
	})

	requestprocessors = requestprocessors.DeepCopy()
	cli.SortByNamespaceAndName(requestprocessors.Items)

	return tablePrinter.PrintObj(requestprocessors, c.Stdout)
}

func NewRequestProcessorListCommand(c *cli.Config) *cobra.Command {
	opts := &RequestProcessorListOptions{}

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

func printRequestProcessorList(requestprocessors *requestv1alpha1.RequestProcessorList, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(requestprocessors.Items))
	for i := range requestprocessors.Items {
		r, err := printRequestProcessor(&requestprocessors.Items[i], opts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printRequestProcessor(requestprocessor *requestv1alpha1.RequestProcessor, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: requestprocessor},
	}
	row.Cells = append(row.Cells,
		requestprocessor.Name,
		requestProcessorRefType(requestprocessor),
		requestProcessorRef(requestprocessor),
		cli.FormatEmptyString(requestprocessor.Status.Domain),
		cli.FormatConditionStatus(requestprocessor.Status.GetCondition(requestv1alpha1.RequestProcessorConditionReady)),
		cli.FormatTimestampSince(requestprocessor.CreationTimestamp),
	)
	return []metav1beta1.TableRow{row}, nil
}

func printRequestProcessorColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Type", Type: "string"},
		{Name: "Ref", Type: "string"},
		{Name: "Domain", Type: "string"},
		{Name: "Ready", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}

func requestProcessorRefType(requestprocessor *requestv1alpha1.RequestProcessor) string {
	if len(requestprocessor.Spec) == 0 {
		return "<unknown>"
	}
	if requestprocessor.Spec[0].Build == nil {
		return "image"
	}
	if requestprocessor.Spec[0].Build.ApplicationRef != "" {
		return "application"
	}
	if requestprocessor.Spec[0].Build.FunctionRef != "" {
		return "function"
	}
	return "<unknown>"
}

func requestProcessorRef(requestprocessor *requestv1alpha1.RequestProcessor) string {
	if len(requestprocessor.Spec) == 0 {
		return "<unknown>"
	}
	if requestprocessor.Spec[0].Build == nil {
		return requestprocessor.Spec[0].Template.Containers[0].Image
	}
	if requestprocessor.Spec[0].Build.ApplicationRef != "" {
		return requestprocessor.Spec[0].Build.ApplicationRef
	}
	if requestprocessor.Spec[0].Build.FunctionRef != "" {
		return requestprocessor.Spec[0].Build.FunctionRef
	}
	return requestprocessor.Spec[0].Template.Containers[0].Image
}
