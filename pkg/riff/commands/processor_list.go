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
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ProcessorListOptions struct {
	cli.ListOptions
}

var (
	_ cli.Validatable = (*ProcessorListOptions)(nil)
	_ cli.Executable  = (*ProcessorListOptions)(nil)
)

func (opts *ProcessorListOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *ProcessorListOptions) Exec(ctx context.Context, c *cli.Config) error {
	processors, err := c.Stream().Processors(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(processors.Items) == 0 {
		c.Infof("No processors found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := printProcessorColumns()
		h.TableHandler(columns, printProcessorList)
		h.TableHandler(columns, printProcessor)
	})

	processors = processors.DeepCopy()
	cli.SortByNamespaceAndName(processors.Items)

	return tablePrinter.PrintObj(processors, c.Stdout)
}

func NewProcessorListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &ProcessorListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of processors",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s processor list", c.Name),
			fmt.Sprintf("%s processor list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func printProcessorList(processors *streamv1alpha1.ProcessorList, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(processors.Items))
	for i := range processors.Items {
		r, err := printProcessor(&processors.Items[i], opts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printProcessor(processor *streamv1alpha1.Processor, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: processor},
	}
	row.Cells = append(row.Cells,
		processor.Name,
		cli.FormatEmptyString(processor.Spec.FunctionRef),
		cli.FormatEmptyString(strings.Join(processor.Spec.Inputs, ",")),
		cli.FormatEmptyString(strings.Join(processor.Spec.Outputs, ",")),
		cli.FormatConditionStatus(processor.Status.GetCondition(streamv1alpha1.ProcessorConditionReady)),
		cli.FormatTimestampSince(processor.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func printProcessorColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Function", Type: "string"},
		{Name: "Inputs", Type: "string"},
		{Name: "Outputs", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
