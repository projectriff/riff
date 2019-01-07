/*
 * Copyright 2018-2019 The original author or authors
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
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
)

var _ = Describe("The riff system command", func() {

	var (
		systemCommand *cobra.Command
	)

	BeforeEach(func() {
		systemCommand = commands.System()
	})

	It("should be documented", func() {
		Expect(systemCommand.Name()).To(Equal("system"))
		Expect(systemCommand.Short).NotTo(BeEmpty(), "missing short description")
	})

})

var _ = Describe("The riff system install subcommand", func() {

	var (
		client               core.KubectlClient
		clientMock           *mocks.KubectlClient
		systemInstallCommand *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.KubectlClient)
		clientMock = client.(*mocks.KubectlClient)
		systemInstallCommand = commands.SystemInstall(map[string]*core.Manifest{}, &client)
	})

	AfterEach(func() {
		clientMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(systemInstallCommand.Name()).To(Equal("install"))
		Expect(systemInstallCommand.Short).NotTo(BeEmpty(), "missing short description")
		Expect(systemInstallCommand.Long).NotTo(BeEmpty(), "missing long description")
	})

	It("should define flags", func() {
		Expect(systemInstallCommand.Flag("manifest")).NotTo(BeNil())
		Expect(systemInstallCommand.Flag("node-port")).NotTo(BeNil())
		Expect(systemInstallCommand.Flag("force")).NotTo(BeNil())
	})

	Context("when given no args or flags", func() {
		It("should fail", func() {
			systemInstallCommand.SetArgs([]string{})

			err := systemInstallCommand.Execute()

			Expect(err).To(MatchError("required flag(s) \"manifest\" not set"))
		})
	})

	Context("when given wrong args or flags", func() {
		It("should fail with any argument", func() {
			systemInstallCommand.SetArgs([]string{"extraneous-arg"})

			err := systemInstallCommand.Execute()

			Expect(err).To(MatchError("accepts 0 arg(s), received 1"))
		})
	})

	Context("when given valid args and flags", func() {
		It("should complete the install", func() {
			completed := true
			stdout := &strings.Builder{}
			systemInstallCommand.SetOutput(stdout)
			clientMock.On("SystemInstall",
				map[string]*core.Manifest{},
				core.SystemInstallOptions{
					Manifest: "some/path.yaml",
				}).
				Return(completed, nil)

			systemInstallCommand.SetArgs([]string{"-m", "some/path.yaml"})
			err := systemInstallCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("\ninstall completed successfully\n"))
		})

		It("should report the install interruption", func() {
			completed := false
			stdout := &strings.Builder{}
			systemInstallCommand.SetOutput(stdout)
			clientMock.On("SystemInstall",
				map[string]*core.Manifest{},
				core.SystemInstallOptions{
					Manifest: "some/path.yaml",
				}).
				Return(completed, nil)

			systemInstallCommand.SetArgs([]string{"-m", "some/path.yaml"})
			err := systemInstallCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("\ninstall was interrupted\n"))
		})

		It("should propagate client error", func() {
			expectedError := errors.New("client failure")
			clientMock.On("SystemInstall",
				map[string]*core.Manifest{},
				core.SystemInstallOptions{
					Manifest: "some/path.yaml",
				}).
				Return(false, expectedError)

			systemInstallCommand.SetArgs([]string{"-m", "some/path.yaml"})
			err := systemInstallCommand.Execute()

			Expect(err).To(MatchError(expectedError))
		})
	})

})

var _ = Describe("The riff system uninstall subcommand", func() {

	var (
		client                 core.KubectlClient
		clientMock             *mocks.KubectlClient
		systemUninstallCommand *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.KubectlClient)
		clientMock = client.(*mocks.KubectlClient)
		systemUninstallCommand = commands.SystemUninstall(&client)
	})

	AfterEach(func() {
		clientMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(systemUninstallCommand.Name()).To(Equal("uninstall"))
		Expect(systemUninstallCommand.Short).NotTo(BeEmpty(), "missing short description")
		Expect(systemUninstallCommand.Long).NotTo(BeEmpty(), "missing long description")
		Expect(systemUninstallCommand.Example).NotTo(BeEmpty(), "missing long description")
	})

	It("should define flags", func() {
		Expect(systemUninstallCommand.Flag("istio")).NotTo(BeNil())
		Expect(systemUninstallCommand.Flag("force")).NotTo(BeNil())
	})

	Context("when given wrong args or flags", func() {
		It("should fail with any argument", func() {
			systemUninstallCommand.SetArgs([]string{"extraneous-arg"})

			err := systemUninstallCommand.Execute()

			Expect(err).To(MatchError("accepts 0 arg(s), received 1"))
		})
	})

	Context("when given valid args and flags", func() {
		It("should complete the uninstall", func() {
			completed := true
			stdout := &strings.Builder{}
			systemUninstallCommand.SetOutput(stdout)
			clientMock.On("SystemUninstall",
				core.SystemUninstallOptions{}).
				Return(completed, nil)

			err := systemUninstallCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("\nuninstall completed successfully\n"))
		})

		It("should report the uninstall interruption", func() {
			completed := false
			stdout := &strings.Builder{}
			systemUninstallCommand.SetOutput(stdout)
			clientMock.On("SystemUninstall",
				core.SystemUninstallOptions{}).
				Return(completed, nil)

			err := systemUninstallCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("\nuninstall was interrupted\n"))
		})

		It("should propagate client error", func() {
			expectedError := errors.New("client failure")
			clientMock.On("SystemUninstall",
				core.SystemUninstallOptions{}).
				Return(false, expectedError)

			err := systemUninstallCommand.Execute()

			Expect(err).To(MatchError(expectedError))
		})
	})

})
