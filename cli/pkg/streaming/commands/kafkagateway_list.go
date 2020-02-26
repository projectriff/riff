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

type KafkaGatewayListOptions struct {
	options.ListOptions
}

var (
	_ cli.Validatable = (*KafkaGatewayListOptions)(nil)
	_ cli.Executable  = (*KafkaGatewayListOptions)(nil)
)

func (opts *KafkaGatewayListOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *KafkaGatewayListOptions) Exec(ctx context.Context, c *cli.Config) error {
	gateways, err := c.StreamingRuntime().KafkaGateways(opts.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(gateways.Items) == 0 {
		c.Infof("No kafka gateways found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	gateways = gateways.DeepCopy()
	cli.SortByNamespaceAndName(gateways.Items)

	return tablePrinter.PrintObj(gateways, c.Stdout)
}

func NewKafkaGatewayListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &KafkaGatewayListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of kafka gateways",
		Long: strings.TrimSpace(`
List Kafka gateways in a namespace or across all namespaces.

For detail regarding the status of a single Kafka gateway, run:

    ` + c.Name + ` streaming kafka-gateway status <kafka-gateway-name>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s streaming kafka-gateway list", c.Name),
			fmt.Sprintf("%s streaming kafka-gateway list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func (opts *KafkaGatewayListOptions) printList(gateways *streamv1alpha1.KafkaGatewayList, printOpts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(gateways.Items))
	for i := range gateways.Items {
		r, err := opts.print(&gateways.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *KafkaGatewayListOptions) print(gateway *streamv1alpha1.KafkaGateway, _ printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: gateway},
	}
	row.Cells = append(row.Cells,
		gateway.Name,
		cli.FormatEmptyString(gateway.Spec.BootstrapServers),
		cli.FormatConditionStatus(gateway.Status.GetCondition(streamv1alpha1.KafkaGatewayConditionReady)),
		cli.FormatTimestampSince(gateway.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *KafkaGatewayListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Bootstrap Servers", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
