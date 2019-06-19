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
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/spf13/cobra"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FunctionStatusOptions struct {
	cli.ResourceOptions
}

var (
	_ cli.Validatable = (*FunctionStatusOptions)(nil)
	_ cli.Executable  = (*FunctionStatusOptions)(nil)
)

func (opts *FunctionStatusOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	return errs
}

func (opts *FunctionStatusOptions) Exec(ctx context.Context, c *cli.Config) error {
	function, err := c.Build().Functions(opts.Namespace).Get(opts.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}
		c.Errorf("Function %q not found\n",
			fmt.Sprintf("%s/%s", opts.Namespace, opts.Name))
		return err
	}
	ready := function.Status.GetCondition(buildv1alpha1.FunctionConditionReady)
	cli.PrintResourceStatus(c, function.Name, ready)
	return nil
}

func NewFunctionStatusCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &FunctionStatusOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "show function status",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s function status my-function", c.Name),
		}, "\n"),
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)

	return cmd
}
