/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"os"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	invoker "github.com/projectriff/riff/riff-cli/pkg/invoker"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
)

var _ = Describe("The invokers commands", func() {
	var (
		normalKubeCtl *kubectl.MockKubeCtl

		rootCommand           *cobra.Command
		invokersCommand       *cobra.Command
		invokersApplyCommand  *cobra.Command
		invokersListCommand   *cobra.Command
		invokersDeleteCommand *cobra.Command

		invokersApplyOptions  *invoker.ApplyOptions
		invokersDeleteOptions *InvokersDeleteOptions

		oldCWD string
	)

	invokerWithNameAndVersion := func(name string, version string) func(stdin *[]byte) bool {
		return func(stdin *[]uint8) bool {
			invoker := projectriff_v1.Invoker{}
			err := yaml.Unmarshal(*stdin, &invoker)
			if err != nil {
				return false
			}
			return invoker.ObjectMeta.Name == name && invoker.Spec.Version == version
		}
	}

	BeforeEach(func() {
		var err error

		normalKubeCtl = new(kubectl.MockKubeCtl)

		rootCommand = Root()
		invokersCommand = Invokers()
		invokersApplyCommand, invokersApplyOptions = InvokersApply(normalKubeCtl)
		invokersListCommand = InvokersList(normalKubeCtl)
		invokersDeleteCommand, invokersDeleteOptions = InvokersDelete(normalKubeCtl)

		rootCommand.AddCommand(invokersCommand)
		invokersCommand.AddCommand(
			invokersApplyCommand,
			invokersListCommand,
			invokersDeleteCommand,
		)

		oldCWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		normalKubeCtl.AssertExpectations(GinkgoT())

		os.Chdir(oldCWD)
	})

	Context("invokers apply", func() {

		It("applies invokers from the current directory", func() {
			os.Chdir("../test_data/invokers")

			rootCommand.SetArgs([]string{"invokers", "apply"})

			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("command", "0.0.5-snapshot"))).Return("", nil).Once()
			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("java", "0.0.5-snapshot"))).Return("", nil).Once()
			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("node", "0.0.5-snapshot"))).Return("", nil).Once()
			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("python2", "0.0.5-snapshot"))).Return("", nil).Once()
			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("python3", "0.0.5-snapshot"))).Return("", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal("."))
		})

		It("applies a specific invoker from the current directory", func() {
			os.Chdir("../test_data/invokers")

			rootCommand.SetArgs([]string{"invokers", "apply", "-f", "command-function-invoker.yaml"})

			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("command", "0.0.5-snapshot"))).Return("", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal("command-function-invoker.yaml"))
		})

		It("applies a specific invoker from another directory", func() {
			rootCommand.SetArgs([]string{"invokers", "apply", "-f", "../test_data/invokers/command-function-invoker.yaml"})

			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("command", "0.0.5-snapshot"))).Return("", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal("../test_data/invokers/command-function-invoker.yaml"))
		})

		It("applies a specific invoker with a custom version", func() {
			rootCommand.SetArgs([]string{"invokers", "apply", "-v", "latest", "-f", "../test_data/invokers/command-function-invoker.yaml"})

			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("command", "latest"))).Return("", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal("../test_data/invokers/command-function-invoker.yaml"))
			Expect(invokersApplyOptions.Version).To(Equal("latest"))
		})

		It("applies a specific invoker with a custom name", func() {
			rootCommand.SetArgs([]string{"invokers", "apply", "-n", "foobar", "-f", "../test_data/invokers/command-function-invoker.yaml"})

			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("foobar", "0.0.5-snapshot"))).Return("", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal("../test_data/invokers/command-function-invoker.yaml"))
			Expect(invokersApplyOptions.Name).To(Equal("foobar"))
		})

		It("fails to apply multiple invokers with a custom name", func() {
			os.Chdir("../test_data/invokers")

			rootCommand.SetArgs([]string{"invokers", "apply", "-n", "foobar"})

			err := rootCommand.Execute()
			Expect(err).To(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal("."))
			Expect(invokersApplyOptions.Name).To(Equal("foobar"))
		})

		// NOTE this test requires network access to GitHub and may break if history is rewritten
		It("applies a specific invoker from a URL", func() {
			goodURL := "https://github.com/projectriff/riff/raw/b81e42181f46bfcf7cb2e230485422be2d3807d3/riff-cli/test_data/invokers/command-function-invoker.yaml"

			rootCommand.SetArgs([]string{"invokers", "apply", "-f", goodURL})

			normalKubeCtl.On("ExecStdin", []string{"apply", "-f", "-"}, mock.MatchedBy(invokerWithNameAndVersion("command", "0.0.5-snapshot"))).Return("", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal(goodURL))
		})

		It("fails to apply a specific invoker from a bogus filepath", func() {
			rootCommand.SetArgs([]string{"invokers", "apply", "-f", "bogus-invoker.yaml"})

			err := rootCommand.Execute()
			Expect(err).To(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal("bogus-invoker.yaml"))
		})

		It("fails to apply a specific invoker from a bogus URL", func() {
			badURL := "https://github.com/projectriff/riff/raw/b81e42181f46bfcf7cb2e230485422be2d3807d3/riff-cli/test_data/invokers/bogus-invoker.yaml"
			rootCommand.SetArgs([]string{"invokers", "apply", "-f", badURL})

			err := rootCommand.Execute()
			Expect(err).To(HaveOccurred())

			Expect(invokersApplyOptions.Filename).To(Equal(badURL))
		})

	})

	Context("invokers list", func() {

		It("lists installed invokers", func() {
			rootCommand.SetArgs([]string{"invokers", "list"})

			normalKubeCtl.On(
				"Exec",
				[]string{
					"get", "invokers.projectriff.io",
					"--sort-by=metadata.name",
					"-o=custom-columns=INVOKER:.metadata.name,VERSION:.spec.version",
				},
			).Return("INVOKER      VERSION\n", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())
		})

	})

	Context("invokers delete", func() {

		It("removes the specified invoker", func() {
			rootCommand.SetArgs([]string{"invokers", "delete", "command"})

			normalKubeCtl.On("Exec", []string{"delete", "invokers.projectriff.io", "command"}).Return("", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(invokersDeleteOptions.All).To(BeFalse())
			Expect(invokersDeleteOptions.Name).To(Equal("command"))
		})

		It("removes all invokers", func() {
			rootCommand.SetArgs([]string{"invokers", "delete", "--all"})

			normalKubeCtl.On("Exec", []string{"delete", "invokers.projectriff.io", "--all"}).Return("", nil).Once()

			err := rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(invokersDeleteOptions.All).To(BeTrue())
			Expect(invokersDeleteOptions.Name).To(Equal(""))
		})

	})

})
