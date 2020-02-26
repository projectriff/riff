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

package commands_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/projectriff/cli/pkg/build/commands"
	"github.com/projectriff/cli/pkg/cli"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	kailtesting "github.com/projectriff/cli/pkg/testing/kail"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFunctionTailOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.FunctionTailOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.FunctionTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "since duration",
			Options: &commands.FunctionTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Since:           "1m",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid duration",
			Options: &commands.FunctionTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Since:           "1",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("1", cli.SinceFlagName),
		},
	}

	table.Run(t)
}

func TestFunctionTailCommand(t *testing.T) {
	defaultNamespace := "default"
	functionName := "my-function"
	function := &buildv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      functionName,
		},
	}

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "show logs",
			Args: []string{functionName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("FunctionLogs", mock.Anything, function, cli.TailSinceDefault, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...log output...\n")
				})
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				function,
			},
			ExpectOutput: `
...log output...
`,
		},
		{
			Name: "show logs since",
			Args: []string{functionName, cli.SinceFlagName, "1h"},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("FunctionLogs", mock.Anything, function, time.Hour, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...log output...\n")
				})
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				function,
			},
			ExpectOutput: `
...log output...
`,
		},
		{
			Name:        "unknown function",
			Args:        []string{functionName},
			ShouldError: true,
		},
		{
			Name: "kail error",
			Args: []string{functionName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("FunctionLogs", mock.Anything, function, cli.TailSinceDefault, mock.Anything).Return(fmt.Errorf("kail error"))
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				function,
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewFunctionTailCommand)
}
