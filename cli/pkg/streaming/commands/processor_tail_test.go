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

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/streaming/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	kailtesting "github.com/projectriff/cli/pkg/testing/kail"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestProcessorTailOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.ProcessorTailOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.ProcessorTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "since duration",
			Options: &commands.ProcessorTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Since:           "1m",
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid duration",
			Options: &commands.ProcessorTailOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Since:           "1",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("1", cli.SinceFlagName),
		},
	}

	table.Run(t)
}

func TestProcessorTailCommand(t *testing.T) {
	defaultNamespace := "default"
	processorName := "my-processor"
	processor := &streamv1alpha1.Processor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      processorName,
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
			Args: []string{processorName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("StreamingProcessorLogs", mock.Anything, processor, cli.TailSinceDefault, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
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
				processor,
			},
			ExpectOutput: `
...log output...
`,
		},
		{
			Name: "show logs since",
			Args: []string{processorName, cli.SinceFlagName, "1h"},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("StreamingProcessorLogs", mock.Anything, processor, time.Hour, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
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
				processor,
			},
			ExpectOutput: `
...log output...
`,
		},
		{
			Name:        "unknown processor",
			Args:        []string{processorName},
			ShouldError: true,
		},
		{
			Name: "kail error",
			Args: []string{processorName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("StreamingProcessorLogs", mock.Anything, processor, cli.TailSinceDefault, mock.Anything).Return(fmt.Errorf("kail error"))
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				processor,
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewProcessorTailCommand)
}
