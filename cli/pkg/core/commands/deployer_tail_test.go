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

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/core/commands"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	kailtesting "github.com/projectriff/riff/cli/pkg/testing/kail"
	corev1alpha1 "github.com/projectriff/system/pkg/apis/core/v1alpha1"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDeployerTailOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.DeployerTailOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.DeployerTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "since duration",
			Options: &commands.DeployerTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Since:           "1m",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid duration",
			Options: &commands.DeployerTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Since:           "1",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("1", cli.SinceFlagName),
		},
	}

	table.Run(t)
}

func TestDeployerTailCommand(t *testing.T) {
	defaultNamespace := "default"
	deployerName := "my-deployer"
	deployer := &corev1alpha1.Deployer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      deployerName,
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
			Args: []string{deployerName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("CoreDeployerLogs", mock.Anything, deployer, cli.TailSinceDefault, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
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
				deployer,
			},
			ExpectOutput: `
...log output...
`,
		},
		{
			Name: "show logs since",
			Args: []string{deployerName, cli.SinceFlagName, "1h"},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("CoreDeployerLogs", mock.Anything, deployer, time.Hour, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
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
				deployer,
			},
			ExpectOutput: `
...log output...
`,
		},
		{
			Name:        "unknown deployer",
			Args:        []string{deployerName},
			ShouldError: true,
		},
		{
			Name: "kail error",
			Args: []string{deployerName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("CoreDeployerLogs", mock.Anything, deployer, cli.TailSinceDefault, mock.Anything).Return(fmt.Errorf("kail error"))
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				deployer,
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewDeployerTailCommand)
}
