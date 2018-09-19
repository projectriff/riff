/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package commands_test

import (
	"fmt"

	"strings"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("The riff function create command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			fc         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			fc = commands.FunctionCreate(&mockClient)
		})
		It("should fail with no args", func() {
			fc.SetArgs([]string{})
			err := fc.Execute()
			Expect(err).To(MatchError("accepts 2 arg(s), received 0"))
		})
		It("should fail with invalid invoker or function name", func() {
			fc.SetArgs([]string{".invalid", "fn-name"})
			err := fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))

			fc = commands.FunctionCreate(&mockClient)
			fc.SetArgs([]string{"node", "invålid"})
			err = fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
		It("should fail without required flags", func() {
			fc.SetArgs([]string{"node", "square", "--local-path", "."})
			err := fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("required flag(s)")))
			Expect(err).To(MatchError(ContainSubstring("image")))
		})
		It("should fail without required source location flags", func() {
			fc.SetArgs([]string{"node", "square"})
			err := fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("at least one of")))
			Expect(err).To(MatchError(ContainSubstring("--git-repo")))
			Expect(err).To(MatchError(ContainSubstring("--local-path")))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			fc     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			fc = commands.FunctionCreate(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())

		})
		It("should involve the core.Client", func() {
			fc.SetArgs([]string{"node", "square", "--image", "foo/bar", "--git-repo", "https://github.com/repo"})

			o := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				InvokerURL:  "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
			}
			o.Name = "square"
			o.Image = "foo/bar"
			o.Env = []string{}
			o.EnvFrom = []string{}

			asMock.On("CreateFunction", o, mock.Anything).Return(nil, nil)
			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			fc.SetArgs([]string{"node", "square", "--image", "foo/bar", "--git-repo", "https://github.com/repo"})

			e := fmt.Errorf("some error")
			asMock.On("CreateFunction", mock.Anything, mock.Anything).Return(nil, e)
			err := fc.Execute()
			Expect(err).To(MatchError(e))
		})
		It("should add env vars when asked to", func() {
			fc.SetArgs([]string{"node", "square", "--image", "foo/bar", "--git-repo", "https://github.com/repo",
				"--env", "FOO=bar", "--env", "BAZ=qux", "--env-from", "secretKeyRef:foo:bar"})

			o := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				InvokerURL:  "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
			}
			o.Name = "square"
			o.Image = "foo/bar"
			o.Env = []string{"FOO=bar", "BAZ=qux"}
			o.EnvFrom = []string{"secretKeyRef:foo:bar"}

			asMock.On("CreateFunction", o, mock.Anything).Return(nil, nil)
			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should print when --dry-run is set", func() {
			fc.SetArgs([]string{"node", "square", "--image", "foo/bar", "--git-repo", "https://github.com/repo", "--dry-run"})

			functionOptions := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				InvokerURL:  "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
			}
			functionOptions.Name = "square"
			functionOptions.Image = "foo/bar"
			functionOptions.Env = []string{}
			functionOptions.EnvFrom = []string{}
			functionOptions.DryRun = true

			f := v1alpha1.Service{}
			f.Name = "square"
			asMock.On("CreateFunction", functionOptions, mock.Anything).Return(&f, nil)

			stdout := &strings.Builder{}
			fc.SetOutput(stdout)

			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.String()).To(Equal(fnCreateDryRun))
		})

		It("should display the status hint", func() {
			fc.SetArgs([]string{"node", "square", "--image", "foo/bar", "--git-repo", "https://github.com/repo"})
			functionOptions := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				InvokerURL:  "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
			}
			functionOptions.Name = "square"
			functionOptions.Image = "foo/bar"
			functionOptions.Env = []string{}
			functionOptions.EnvFrom = []string{}
			function := v1alpha1.Service{}
			function.Name = "square"
			asMock.On("CreateFunction", functionOptions, mock.Anything).Return(&function, nil)
			stdout := &strings.Builder{}
			fc.SetOutput(stdout)

			err := fc.Execute()

			Expect(err).NotTo(HaveOccurred())
			fmt.Println(stdout.String())
			Expect(stdout.String()).To(HaveSuffix("Issue `riff service status square` to see the status of the function\n"))
		})

		It("should include the nondefault namespace in the status hint", func() {
			fc.SetArgs([]string{"node", "square", "--image", "foo/bar", "--git-repo", "https://github.com/repo",
				"--namespace", "ns"})
			functionOptions := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				InvokerURL:  "https://github.com/projectriff/node-function-invoker/raw/v0.0.8/node-invoker.yaml",
			}
			functionOptions.Name = "square"
			functionOptions.Namespace = "ns"
			functionOptions.Image = "foo/bar"
			functionOptions.Env = []string{}
			functionOptions.EnvFrom = []string{}
			function := v1alpha1.Service{}
			function.Name = "square"
			function.Namespace = "ns"
			asMock.On("CreateFunction", functionOptions, mock.Anything).Return(&function, nil)
			stdout := &strings.Builder{}
			fc.SetOutput(stdout)

			err := fc.Execute()

			Expect(err).NotTo(HaveOccurred())
			fmt.Println(stdout.String())
			Expect(stdout.String()).To(HaveSuffix("Issue `riff service status square -n ns` to see the status of the function\n"))
		})

	})
})

const fnCreateDryRun = `metadata:
  creationTimestamp: null
  name: square
spec: {}
status: {}
---
`

var _ = Describe("The riff function build command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			fc         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			fc = commands.FunctionBuild(&mockClient)
		})
		It("should fail with no args", func() {
			fc.SetArgs([]string{})
			err := fc.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
		It("should fail with invalid function name", func() {
			//fc = commands.FunctionBuild(&mockClient)
			fc.SetArgs([]string{"invålid"})
			err := fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
	})

	Context("when given suitable args", func() {
		var (
			client core.Client
			asMock *mocks.Client
			fc     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			fc = commands.FunctionBuild(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())

		})
		It("should involve the core.Client", func() {
			fc.SetArgs([]string{"square", "--namespace", "ns"})

			o := core.BuildFunctionOptions{}
			o.Name = "square"
			o.Namespace = "ns"

			asMock.On("BuildFunction", o, mock.Anything).Return(nil)
			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			fc.SetArgs([]string{"square", "--namespace", "ns"})

			e := fmt.Errorf("some error")
			asMock.On("BuildFunction", mock.Anything, mock.Anything).Return(e)
			err := fc.Execute()
			Expect(err).To(MatchError(e))
		})
	})
})
