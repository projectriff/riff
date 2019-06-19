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

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/cli/printers"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type DoctorOptions struct {
}

var (
	_ cli.Validatable = (*DoctorOptions)(nil)
	_ cli.Executable  = (*DoctorOptions)(nil)
)

func (opts *DoctorOptions) Validate(ctx context.Context) *cli.FieldError {
	return cli.EmptyFieldError
}

func (opts *DoctorOptions) Exec(ctx context.Context, c *cli.Config) error {
	ok, err := opts.checkNamespaces(c)
	if err != nil {
		return err
	}
	c.Printf("\n")
	if ok {
		c.Successf("Installation is OK\n")
	} else {
		c.Errorf("Installation is not healthy\n")
	}
	return nil
}

func (*DoctorOptions) checkNamespaces(c *cli.Config) (bool, error) {
	namespaces, err := c.Core().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	foundNamespaces := sets.NewString()
	for _, namespace := range namespaces.Items {
		foundNamespaces.Insert(namespace.Name)
	}
	requiredNamespaces := []string{
		"istio-system",
		"knative-build",
		"knative-serving",
		"riff-system",
	}
	printer := printers.GetNewTabWriter(c.Stdout)
	defer printer.Flush()
	ok := true
	for _, namespace := range requiredNamespaces {
		var status string
		if foundNamespaces.Has(namespace) {
			status = cli.Ssuccessf("OK")
		} else {
			ok = false
			status = cli.Serrorf("Missing")
		}
		fmt.Fprintf(printer, "Namespace %q\t%s\n", namespace, status)
	}
	return ok, nil
}

func NewDoctorCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &DoctorOptions{}

	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"doc"},
		Short:   "check riff's requirements are installed",
		Long: strings.TrimSpace(`
Check riff's requirements are installed

1. check namespaces are present in Kubernetes
istio-system
knative-build
knative-serving
riff-system
    `),
		Example: "riff doctor",
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	return cmd
}
