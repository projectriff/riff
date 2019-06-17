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

type StreamListOptions struct {
	cli.ListOptions
}

var (
	_ cli.Validatable = (*StreamListOptions)(nil)
	_ cli.Executable  = (*StreamListOptions)(nil)
)

func (opts *StreamListOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *StreamListOptions) Exec(ctx context.Context, c *cli.Config) error {
	streams, err := c.Stream().Streams(opts.Namespace).List(metav1.ListOptions{})
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
		columns := printStreamColumns()
		h.TableHandler(columns, printStreamList)
		h.TableHandler(columns, printStream)
	})

	streams = streams.DeepCopy()
	cli.SortByNamespaceAndName(streams.Items)

	return tablePrinter.PrintObj(streams, c.Stdout)
}

func NewStreamListCommand(c *cli.Config) *cobra.Command {
	opts := &StreamListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of streams",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s stream list", c.Name),
			fmt.Sprintf("%s stream list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(c, opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func printStreamList(streams *streamv1alpha1.StreamList, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(streams.Items))
	for i := range streams.Items {
		r, err := printStream(&streams.Items[i], opts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printStream(stream *streamv1alpha1.Stream, opts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: stream},
	}
	row.Cells = append(row.Cells,
		stream.Name,
		cli.FormatEmptyString(stream.Status.Address.Topic),
		cli.FormatEmptyString(stream.Status.Address.Gateway),
		cli.FormatEmptyString(stream.Spec.Provider),
		cli.FormatEmptyString(stream.Spec.ContentType),
		cli.FormatConditionStatus(stream.Status.GetCondition(streamv1alpha1.StreamConditionReady)),
		cli.FormatTimestampSince(stream.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func printStreamColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Topic", Type: "string"},
		{Name: "Gateway", Type: "string"},
		{Name: "Provider", Type: "string"},
		{Name: "Content-Type", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
