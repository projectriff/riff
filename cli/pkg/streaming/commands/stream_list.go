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
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type StreamListOptions struct {
	options.ListOptions
}

var (
	_ cli.Validatable = (*StreamListOptions)(nil)
	_ cli.Executable  = (*StreamListOptions)(nil)
)

func (opts *StreamListOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *StreamListOptions) Exec(ctx context.Context, c *cli.Config) error {
	streams, err := c.StreamingRuntime().Streams(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(streams.Items) == 0 {
		c.Infof("No streams found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	streams = streams.DeepCopy()
	cli.SortByNamespaceAndName(streams.Items)

	return tablePrinter.PrintObj(streams, c.Stdout)
}

func NewStreamListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &StreamListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of streams",
		Long: strings.TrimSpace(`
List streams in a namespace or across all namespaces.

For detail regarding the status of a single stream, run:

    ` + c.Name + ` streaming stream status <stream-name>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s streaming stream list", c.Name),
			fmt.Sprintf("%s streaming stream list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func (opts *StreamListOptions) printList(streams *streamv1alpha1.StreamList, printOpts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(streams.Items))
	for i := range streams.Items {
		r, err := opts.print(&streams.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *StreamListOptions) print(stream *streamv1alpha1.Stream, _ printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: stream},
	}
	row.Cells = append(row.Cells,
		stream.Name,
		cli.FormatEmptyString(stream.Spec.Gateway.Name),
		cli.FormatEmptyString(stream.Spec.ContentType),
		cli.FormatConditionStatus(stream.Status.GetCondition(streamv1alpha1.StreamConditionReady)),
		cli.FormatTimestampSince(stream.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *StreamListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Gateway", Type: "string"},
		{Name: "Content-Type", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
