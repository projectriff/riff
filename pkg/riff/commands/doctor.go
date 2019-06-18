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
	missingNamespaces, err := opts.checkMissingNamespaces(c)
	if err != nil {
		return err
	}
	if len(missingNamespaces) > 0 {
		msg := "Something is wrong!\n"
		for _, namespace := range missingNamespaces {
			msg += fmt.Sprintf("missing %s\n", namespace)
		}
		c.Errorf(msg)
	} else {
		c.Successf("Installation is OK\n")
	}
	return nil
}

func (*DoctorOptions) checkMissingNamespaces(c *cli.Config) ([]string, error) {
	namespaces, err := c.Core().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	foundNamespaces := sets.NewString()
	for _, namespace := range namespaces.Items {
		foundNamespaces.Insert(namespace.Name)
	}
	requiredNamespaces := sets.NewString(
		"istio-system",
		"knative-build",
		"knative-serving",
		"riff-system",
	)
	missingNamespaces := requiredNamespaces.Difference(foundNamespaces)
	return missingNamespaces.List(), nil
}

func NewDoctorCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &DoctorOptions{}

	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"doc"},
		Short:   "check riff's requirements are installed",
		Long: strings.TrimSpace(`
    <todo>
    `),
		Example: "riff doctor",
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	return cmd
}
