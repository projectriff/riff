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
)

type DoctorOptions struct {
}

var RequiredNamespaces = []string{
	"istio-system",
	"knative-build",
	"knative-serving",
	"riff-system",
}

func (opts *DoctorOptions) Validate(ctx context.Context) *cli.FieldError {
	return cli.EmptyFieldError
}

func (opts *DoctorOptions) Exec(ctx context.Context, c *cli.Config) error {
	missingNamespaces, err := checkMissingNamespaces(c)

	if err != nil {
		c.Errorf(err.Error())
		return err
	} else if len(missingNamespaces) > 0 {
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

func NewDoctorCommand(c *cli.Config) *cobra.Command {
	opts := &DoctorOptions{}

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "check riff's requirements are installed",
		Long: strings.TrimSpace(`
    <todo>
    `),
		Example: "riff doctor",
		Args:    cli.Args(),
		PreRunE: cli.ValidateOptions(opts),
		RunE:    cli.ExecOptions(c, opts),
	}

	return cmd
}

func checkMissingNamespaces(c *cli.Config) ([]string, error) {
	missingNamespaces := []string{}
	namespaces, err := c.Core().Namespaces().List(metav1.ListOptions{})

	if namespaces == nil || len(namespaces.Items) == 0 {
		missingNamespaces = RequiredNamespaces
	} else {
		names := []string{}
		for _, namespace := range namespaces.Items {
			names = append(names, namespace.Name)
		}
		for _, requiredNamespace := range RequiredNamespaces {
			if !stringInSlice(requiredNamespace, names) {
				missingNamespaces = append(missingNamespaces, requiredNamespace)
			}
		}
	}

	return missingNamespaces, err
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
