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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type CredentialListOptions struct {
	options.ListOptions
}

var (
	_ cli.Validatable = (*CredentialListOptions)(nil)
	_ cli.Executable  = (*CredentialListOptions)(nil)
)

func (opts *CredentialListOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	errs = errs.Also(opts.ListOptions.Validate(ctx))

	return errs
}

func (opts *CredentialListOptions) Exec(ctx context.Context, c *cli.Config) error {
	secrets, err := c.Core().Secrets(opts.Namespace).List(metav1.ListOptions{
		LabelSelector: buildv1alpha1.CredentialLabelKey,
	})
	if err != nil {
		return err
	}

	if len(secrets.Items) == 0 {
		c.Infof("No credentials found.\n")
		return nil
	}

	tablePrinter := printers.NewTablePrinter(printers.PrintOptions{
		WithNamespace: opts.AllNamespaces,
	}).With(func(h printers.PrintHandler) {
		columns := opts.printColumns()
		h.TableHandler(columns, opts.printList)
		h.TableHandler(columns, opts.print)
	})

	secrets = secrets.DeepCopy()
	cli.SortByNamespaceAndName(secrets.Items)

	return tablePrinter.PrintObj(secrets, c.Stdout)
}

func NewCredentialListCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &CredentialListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "table listing of credentials",
		Long: strings.TrimSpace(`
List credentials in a namespace or across all namespaces.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s credential list", c.Name),
			fmt.Sprintf("%s credential list %s", c.Name, cli.AllNamespacesFlagName),
		}, "\n"),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.AllNamespacesFlag(cmd, c, &opts.Namespace, &opts.AllNamespaces)

	return cmd
}

func (opts *CredentialListOptions) printList(credentials *corev1.SecretList, printOpts printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(credentials.Items))
	for i := range credentials.Items {
		r, err := opts.print(&credentials.Items[i], printOpts)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func (opts *CredentialListOptions) print(credential *corev1.Secret, _ printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	now := time.Now()
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: credential.DeepCopy()},
	}
	row.Cells = append(row.Cells,
		credential.Name,
		credential.Labels[buildv1alpha1.CredentialLabelKey],
		credential.Annotations["build.pivotal.io/docker"],
		cli.FormatTimestampSince(credential.CreationTimestamp, now),
	)
	return []metav1beta1.TableRow{row}, nil
}

func (opts *CredentialListOptions) printColumns() []metav1beta1.TableColumnDefinition {
	return []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Type", Type: "string"},
		{Name: "Registry", Type: "string"},
		{Name: "Age", Type: "string"},
	}
}
