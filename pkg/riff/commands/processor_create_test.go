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
	"fmt"
	"testing"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	kailtesting "github.com/projectriff/riff/pkg/testing/kail"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestProcessorCreateOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldError: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingField(cli.FunctionRefFlagName),
				cli.ErrMissingField(cli.InputFlagName),
			),
		},
		{
			Name: "with inputs but no outputs",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input1", "input2"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with inputs and outputs",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input1", "input2"},
				Outputs:         []string{"output1", "output2"},
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestProcessorCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	processorName := "my-processor"
	functionRef := "my-func"
	inputName := "input"
	outputName := "output"
	inputNameOther := "otherinput"
	outputNameOther := "otheroutput"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "create with single input",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "create with multiple inputs",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.InputFlagName, inputNameOther},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName, inputNameOther},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "create with single output",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.InputFlagName, inputNameOther, cli.OutputFlagName, outputName},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName, inputNameOther},
						Outputs:     []string{outputName},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "create with multiple outputs",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.InputFlagName, inputNameOther, cli.OutputFlagName, outputName, cli.OutputFlagName, outputNameOther},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName, inputNameOther},
						Outputs:     []string{outputName, outputNameOther},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "error existing processor",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "processors"),
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "tail logs",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.TailFlagName},
			Prepare: func(t *testing.T, c *cli.Config) error {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("ProcessorLogs", mock.Anything, &streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName},
						Outputs:     []string{},
					},
				}, cli.TailSinceCreateDefault, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...log output...\n")
				})
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName},
						Outputs:     []string{},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
Error: WaitUntilReady not implemented for tests
...log output...
`,
		},
		{
			Name: "tail error",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.TailFlagName},
			Prepare: func(t *testing.T, c *cli.Config) error {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("ProcessorLogs", mock.Anything, &streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName},
						Outputs:     []string{},
					},
				}, cli.TailSinceCreateDefault, mock.Anything).Return(fmt.Errorf("kail error"))
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						FunctionRef: functionRef,
						Inputs:      []string{inputName},
						Outputs:     []string{},
					},
				},
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewProcessorCreateCommand)
}
