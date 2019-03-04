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
 */

package commands_test

import (
	"fmt"
	"strings"

	"github.com/projectriff/riff/pkg/core/mocks/mockbuilder"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("The riff function create command", func() {
	Context("when given wrong args or flags", func() {
		var (
			builder core.Builder
			client  core.Client
			asMock  *mocks.Client
			fc      *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			builder = nil
			asMock = client.(*mocks.Client)
			defaults := func() (*core.PackDefaults, error) {
				return &core.PackDefaults{
					BuilderImage: "projectriff/builder",
					RunImage:     "packs/run",
				}, nil
			}

			fc = commands.FunctionCreate(builder, &client, defaults)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())
		})
		It("should fail with no args", func() {
			fc.SetArgs([]string{})
			err := fc.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
		It("should fail with invalid function name", func() {
			fc.SetArgs([]string{".invalid"})
			err := fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
		It("should fail without required flags", func() {
			fc.SetArgs([]string{"square", "--local-path", "."})
			asMock.On("DefaultBuildImagePrefix", "").Return("", nil)
			err := fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("required flag(s)")))
			Expect(err).To(MatchError(ContainSubstring("image")))
		})
		It("should fail without required source location flags", func() {
			fc.SetArgs([]string{"square"})
			err := fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("at least one of")))
			Expect(err).To(MatchError(ContainSubstring("--git-repo")))
			Expect(err).To(MatchError(ContainSubstring("--local-path")))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			builder core.Builder
			client  core.Client
			asMock  *mocks.Client
			fc      *cobra.Command
		)
		BeforeEach(func() {
			builder = new(mockbuilder.Builder)
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)
			defaults := func() (*core.PackDefaults, error) {
				return &core.PackDefaults{
					BuilderImage: "projectriff/builder",
					RunImage:     "packs/run",
				}, nil
			}

			fc = commands.FunctionCreate(builder, &client, defaults)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())
		})
		It("should involve the core.Client", func() {
			fc.SetArgs([]string{"square", "--image", "foo/bar", "--git-repo", "https://github.com/repo"})

			options := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				BuildOptions: core.BuildOptions{
					Invoker:        "",
					BuildpackImage: "projectriff/builder",
					RunImage:       "packs/run",
				},
			}
			options.Name = "square"
			options.Image = "foo/bar"
			options.Env = []string{}
			options.EnvFrom = []string{}

			asMock.On("CreateFunction", builder, options, mock.Anything).Return(nil, nil, nil)
			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should pass correct options from flags", func() {
			fc.SetArgs([]string{"square", "--image", "foo/bar", "--git-repo", "https://github.com/repo", "--invoker", "pascal"})

			options := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				BuildOptions: core.BuildOptions{
					Invoker:        "pascal",
					BuildpackImage: "projectriff/builder",
					RunImage:       "packs/run",
				},
			}
			options.Name = "square"
			options.Image = "foo/bar"
			options.Env = []string{}
			options.EnvFrom = []string{}

			asMock.On("CreateFunction", builder, options, mock.Anything).Return(nil, nil, nil)
			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should provide a default image", func() {
			fc.SetArgs([]string{"square", "--git-repo", "https://github.com/repo"})

			options := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				BuildOptions: core.BuildOptions{
					Invoker:        "",
					BuildpackImage: "projectriff/builder",
					RunImage:       "packs/run",
				},
			}
			options.Name = "square"
			options.Image = "defaulted-prefix/square"
			options.Env = []string{}
			options.EnvFrom = []string{}

			asMock.On("DefaultBuildImagePrefix", "").Return("defaulted-prefix", nil)
			asMock.On("CreateFunction", builder, options, mock.Anything).Return(nil, nil, nil)
			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			fc.SetArgs([]string{"square", "--image", "foo/bar", "--git-repo", "https://github.com/repo"})

			e := fmt.Errorf("some error")
			asMock.On("CreateFunction", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, e)
			err := fc.Execute()
			Expect(err).To(MatchError(e))
		})
		It("should propagate core.Client errors from default image lookup", func() {
			fc.SetArgs([]string{"square", "--git-repo", "https://github.com/repo"})

			e := fmt.Errorf("some error")
			asMock.On("DefaultBuildImagePrefix", "").Return("", e)
			err := fc.Execute()
			Expect(err).To(MatchError("unable to default image: some error"))
		})
		It("should add env vars when asked to", func() {
			fc.SetArgs([]string{"square", "--image", "foo/bar", "--git-repo", "https://github.com/repo",
				"--env", "FOO=bar", "--env", "BAZ=qux", "--env-from", "secretKeyRef:foo:bar"})

			options := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				BuildOptions: core.BuildOptions{
					BuildpackImage: "projectriff/builder",
					RunImage:       "packs/run",
				},
			}
			options.Name = "square"
			options.Image = "foo/bar"
			options.Env = []string{"FOO=bar", "BAZ=qux"}
			options.EnvFrom = []string{"secretKeyRef:foo:bar"}

			asMock.On("CreateFunction", builder, options, mock.Anything).Return(nil, nil, nil)
			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should print when --dry-run is set", func() {
			fc.SetArgs([]string{"square", "--image", "foo/bar", "--git-repo", "https://github.com/repo", "--dry-run"})

			options := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				BuildOptions: core.BuildOptions{
					BuildpackImage: "projectriff/builder",
					RunImage:       "packs/run",
				},
			}
			options.Name = "square"
			options.Image = "foo/bar"
			options.Env = []string{}
			options.EnvFrom = []string{}
			options.DryRun = true

			f := v1alpha1.Service{}
			f.Name = "square"
			cache := corev1.PersistentVolumeClaim{}
			cache.Name = "square-build-cache"
			asMock.On("CreateFunction", builder, options, mock.Anything).Return(&f, &cache, nil)

			stdout := &strings.Builder{}
			fc.SetOutput(stdout)

			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.String()).To(Equal(fnCreateDryRun))
		})

		It("should display the status hint", func() {
			fc.SetArgs([]string{"square", "--image", "foo/bar", "--git-repo", "https://github.com/repo"})
			options := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				BuildOptions: core.BuildOptions{
					BuildpackImage: "projectriff/builder",
					RunImage:       "packs/run",
				},
			}
			options.Name = "square"
			options.Image = "foo/bar"
			options.Env = []string{}
			options.EnvFrom = []string{}
			function := v1alpha1.Service{}
			function.Name = "square"
			cache := corev1.PersistentVolumeClaim{}
			cache.Name = "square-build-cache"
			asMock.On("CreateFunction", builder, options, mock.Anything).Return(&function, &cache, nil)
			stdout := &strings.Builder{}
			fc.SetOutput(stdout)

			err := fc.Execute()

			Expect(err).NotTo(HaveOccurred())
			fmt.Println(stdout.String())
			Expect(stdout.String()).To(HaveSuffix("Issue `riff service status square` to see the status of the function\n"))
		})

		It("should include the nondefault namespace in the status hint", func() {
			fc.SetArgs([]string{"square", "--image", "foo/bar", "--git-repo", "https://github.com/repo",
				"--namespace", "ns"})
			options := core.CreateFunctionOptions{
				GitRepo:     "https://github.com/repo",
				GitRevision: "master",
				BuildOptions: core.BuildOptions{
					BuildpackImage: "projectriff/builder",
					RunImage:       "packs/run",
				},
			}
			options.Name = "square"
			options.Namespace = "ns"
			options.Image = "foo/bar"
			options.Env = []string{}
			options.EnvFrom = []string{}
			function := v1alpha1.Service{}
			function.Name = "square"
			function.Namespace = "ns"
			cache := corev1.PersistentVolumeClaim{}
			cache.Name = "square-build-cache"
			cache.Namespace = "ns"
			asMock.On("CreateFunction", builder, options, mock.Anything).Return(&function, &cache, nil)
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
  name: square-build-cache
spec:
  resources: {}
status: {}
---
metadata:
  creationTimestamp: null
  name: square
spec: {}
status: {}
---
`

var _ = Describe("The riff function update command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockBuilder core.Builder
			mockClient  core.Client
			fc          *cobra.Command
		)
		BeforeEach(func() {
			mockBuilder = nil
			mockClient = nil
			fc = commands.FunctionUpdate(mockBuilder, &mockClient)
		})
		It("should fail with no args", func() {
			fc.SetArgs([]string{})
			err := fc.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
		It("should fail with invalid function name", func() {
			//fc = commands.FunctionUpdate(&mockClient)
			fc.SetArgs([]string{"invålid"})
			err := fc.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
	})

	Context("when given suitable args", func() {
		var (
			builder    core.Builder
			client     core.Client
			clientMock *mocks.Client
			fc         *cobra.Command
		)
		BeforeEach(func() {
			builder = new(mockbuilder.Builder)
			client = new(mocks.Client)
			clientMock = client.(*mocks.Client)

			fc = commands.FunctionUpdate(builder, &client)
		})
		AfterEach(func() {
			clientMock.AssertExpectations(GinkgoT())

		})
		It("should involve the core.Client", func() {
			fc.SetArgs([]string{"square", "--namespace", "ns"})

			options := core.UpdateFunctionOptions{}
			options.Name = "square"
			options.Namespace = "ns"

			clientMock.On("UpdateFunction", builder, options, mock.Anything).Return(nil)
			err := fc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			fc.SetArgs([]string{"square", "--namespace", "ns"})

			e := fmt.Errorf("some error")
			clientMock.On("UpdateFunction", mock.Anything, mock.Anything, mock.Anything).Return(e)
			err := fc.Execute()
			Expect(err).To(MatchError(e))
		})
	})
})
