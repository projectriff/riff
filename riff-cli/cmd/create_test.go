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
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/pkg/docker"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
)

var _ = Describe("The create command", func() {
	var (
		normalDocker *docker.MockDocker
		dryRunDocker *docker.MockDocker

		normalKubeCtl *kubectl.MockKubeCtl
		dryRunKubeCtl *kubectl.MockKubeCtl

		oldCWD string

		commonRiffArgs []string
	)

	BeforeEach(func() {
		var err error

		normalDocker = new(docker.MockDocker)
		dryRunDocker = new(docker.MockDocker)

		normalKubeCtl = new(kubectl.MockKubeCtl)
		dryRunKubeCtl = new(kubectl.MockKubeCtl)

		oldCWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		commonRiffArgs = []string{"--force", "--useraccount", "rifftest"}
	})

	AfterEach(func() {
		normalDocker.AssertExpectations(GinkgoT())
		dryRunDocker.AssertExpectations(GinkgoT())

		normalKubeCtl.AssertExpectations(GinkgoT())
		dryRunKubeCtl.AssertExpectations(GinkgoT())

		os.Chdir(oldCWD)
	})

	Context("init phase scaffolds function", func() {

		It("fails to creates a function with no invokers", func() {
			os.Chdir("../test_data/riff-init/no-invokers")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"create"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("Invokers must be installed, run `riff invokers apply --help` for help")))

			Expect(".").NotTo(HaveUnstagedChanges())
		})

		It("fails to creates a function with an unknown invoker", func() {
			os.Chdir("../test_data/riff-init/no-invokers")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"create", "node"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())

			Expect(".").NotTo(HaveUnstagedChanges())
		})

		It("ignores unknown flags", func() {
			os.Chdir("../test_data/riff-init/no-invokers")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"create", "java", "--handler", "functions.FooFunc"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("Invokers must be installed, run `riff invokers apply --help` for help")))

			Expect(".").NotTo(HaveUnstagedChanges())
		})

		It("creates a function with a matched invoker", func() {
			os.Chdir("../test_data/riff-init/matching-invoker")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"create"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("The invoker must be specified. Pick one of: node")))
		})

		It("creates a function with an explicit invoker", func() {
			os.Chdir("../test_data/riff-init/matching-invoker")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"create", "node"}, commonRiffArgs...))

			normalDocker.On("Exec", "build", "-t", "rifftest/matching-invoker:0.0.1", ".").
				Return(nil).
				Once()
			functionYamlPath, _ := filepath.Abs("matching-invoker-function.yaml")
			topicsYamlPath, _ := filepath.Abs("matching-invoker-topics.yaml")
			linkYamlPath, _ := filepath.Abs("matching-invoker-link.yaml")
			normalKubeCtl.On("Exec", []string{"apply", "-f", functionYamlPath, "-f", topicsYamlPath, "-f", linkYamlPath}).
				Return("function \"matching-invoker\" created\ntopic \"matching-invoker\" created\nlink \"matching-invoker\" created", nil).
				Once()

			err = rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(".").NotTo(HaveUnstagedChanges())
		})

	})

	It("TestCreateCommandImplicitPath", func() {
		invokers, err := stubInvokers("../test_data/invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, initOptions, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		rootCommand.SetArgs(append([]string{"create", "command", "--dry-run", "../test_data/command/echo", "-a", "echo.sh", "-v", "0.0.1-snapshot"}, commonRiffArgs...))

		dryRunDocker.On("Exec", "build", "-t", "rifftest/echo:0.0.1-snapshot", "../test_data/command/echo").
			Return(nil).
			Once()
		topicsYamlPath, _ := filepath.Abs("../test_data/command/echo/echo-topics.yaml")
		dryRunKubeCtl.On("Exec", []string{"apply", "-f", topicsYamlPath}).
			Return("topic \"echo\" created", nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(initOptions.FilePath).To(Equal("../test_data/command/echo"))
		Expect(initOptions.Artifact).To(Equal("echo.sh"))
		Expect(initOptions.UserAccount).To(Equal("rifftest"))
	})

	It("TestCreateCommandFromCWD", func() {
		os.Chdir("../test_data/command/echo")

		invokers, err := stubInvokers("../../invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, _, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		rootCommand.SetArgs(append([]string{"create", "command", "--dry-run", "-a", "echo.sh"}, commonRiffArgs...))

		dryRunDocker.On("Exec", "build", "-t", "rifftest/echo:0.0.1", ".").
			Return(nil).
			Once()
		topicsYamlPath, _ := filepath.Abs("echo-topics.yaml")
		dryRunKubeCtl.On("Exec", []string{"apply", "-f", topicsYamlPath}).
			Return("topic \"echo\" created", nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("TestCreateCommandExplicitPath", func() {
		invokers, err := stubInvokers("../test_data/invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, initOptions, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		path, _ := filepath.Abs("../test_data/command/echo")

		rootCommand.SetArgs(append([]string{"create", "command", "--dry-run", "-f", path, "-v", "0.0.1-snapshot", "-a", "echo.sh"}, commonRiffArgs...))

		dryRunDocker.On("Exec", "build", "-t", "rifftest/echo:0.0.1-snapshot", path).
			Return(nil).
			Once()
		topicsYamlPath := filepath.Join(path, "echo-topics.yaml")
		dryRunKubeCtl.On("Exec", []string{"apply", "-f", topicsYamlPath}).
			Return("topic \"echo\" created", nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(initOptions.FilePath).To(Equal(path))
		Expect(initOptions.Artifact).To(Equal("echo.sh"))
		Expect(initOptions.UserAccount).To(Equal("rifftest"))
	})

	It("TestCreateCommandWithUser", func() {
		invokers, err := stubInvokers("../test_data/invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, initOptions, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		rootCommand.SetArgs([]string{"create", "command", "--dry-run", "../test_data/command/echo", "-u", "me", "-a", "echo.sh"})

		dryRunDocker.On("Exec", "build", "-t", "me/echo:0.0.1", "../test_data/command/echo").
			Return(nil).
			Once()
		topicsYamlPath, _ := filepath.Abs("../test_data/command/echo/echo-topics.yaml")
		dryRunKubeCtl.On("Exec", []string{"apply", "-f", topicsYamlPath}).
			Return("topic \"echo\" created", nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(initOptions.FilePath).To(Equal("../test_data/command/echo"))
		Expect(initOptions.Artifact).To(Equal("echo.sh"))
		Expect(initOptions.UserAccount).To(Equal("me"))
	})

	It("TestCreateCommandExplicitPathAndLang", func() {
		invokers, err := stubInvokers("../test_data/invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, initOptions, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		path, _ := filepath.Abs("../test_data/command/echo")

		rootCommand.SetArgs(append([]string{"create", "command", "--dry-run", path, "-v", "0.0.1-snapshot", "-a", "echo.sh"}, commonRiffArgs...))

		dryRunDocker.On("Exec", "build", "-t", "rifftest/echo:0.0.1-snapshot", path).
			Return(nil).
			Once()
		topicsYamlPath := filepath.Join(path, "echo-topics.yaml")
		dryRunKubeCtl.On("Exec", []string{"apply", "-f", topicsYamlPath}).
			Return("topic \"echo\" created", nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(initOptions.FilePath).To(Equal(path))
		Expect(initOptions.Artifact).To(Equal("echo.sh"))
		Expect(initOptions.UserAccount).To(Equal("rifftest"))
	})

	It("TestCreateLanguageDoesNotMatchArtifact", func() {
		invokers, err := stubInvokers("../test_data/invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, initOptions, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		path := osutils.Path("../test_data/python/demo")

		rootCommand.SetArgs(append([]string{"create", "command", "--dry-run", "-f", path, "-a", "demo.py"}, commonRiffArgs...))

		dryRunDocker.On("Exec", "build", "-t", "rifftest/demo:0.0.1", path).
			Return(nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(initOptions.FilePath).To(Equal(path))
		Expect(initOptions.Artifact).To(Equal("demo.py"))
		Expect(initOptions.UserAccount).To(Equal("rifftest"))
	})

	It("TestCreatePythonCommand", func() {
		invokers, err := stubInvokers("../test_data/invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, initOptions, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		path := osutils.Path("../test_data/python/demo")

		rootCommand.SetArgs(append([]string{"create", "python3", "--dry-run", "-f", path, "-v", "0.0.1-snapshot", "--handler", "process"}, commonRiffArgs...))

		dryRunDocker.On("Exec", "build", "-t", "rifftest/demo:0.0.1-snapshot", path).
			Return(nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(initOptions.FilePath).To(Equal(path))
		Expect(initOptions.Artifact).To(Equal("demo.py"))
		Expect(initOptions.Handler).To(Equal("process"))
		Expect(initOptions.UserAccount).To(Equal("rifftest"))
	})

	It("TestCreatePythonCommandWithDefaultHandler", func() {
		invokers, err := stubInvokers("../test_data/invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, initOptions, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		path := osutils.Path("../test_data/python/demo")

		rootCommand.SetArgs(append([]string{"create", "python3", "--dry-run", "-f", path, "-v", "0.0.1-snapshot"}, commonRiffArgs...))

		dryRunDocker.On("Exec", "build", "-t", "rifftest/demo:0.0.1-snapshot", path).
			Return(nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(initOptions.FilePath).To(Equal(path))
		Expect(initOptions.Artifact).To(Equal("demo.py"))
		Expect(initOptions.Handler).To(Equal("Demo"))
		Expect(initOptions.UserAccount).To(Equal("rifftest"))
	})

	It("TestCreateJavaWithVersion", func() {
		os.Chdir("../test_data/java")

		invokers, err := stubInvokers("../invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, initOptions, _, _, err := setupCreateTest(invokers, normalDocker, dryRunDocker, normalKubeCtl, dryRunKubeCtl)
		Expect(err).NotTo(HaveOccurred())

		rootCommand.SetArgs(append([]string{"create", "java", "--dry-run", "-a", "target/upper-1.0.0.jar", "--handler", "function.Upper"}, commonRiffArgs...))

		dryRunDocker.On("Exec", "build", "-t", "rifftest/java:0.0.1", ".").
			Return(nil).
			Once()

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(initOptions.FilePath).To(Equal("."))
		Expect(initOptions.Artifact).To(Equal("target/upper-1.0.0.jar"))
		Expect(initOptions.Handler).To(Equal("function.Upper"))
		Expect(initOptions.UserAccount).To(Equal("rifftest"))
	})

})

func setupCreateTest(
	invokers []projectriff_v1.Invoker,
	normalDocker *docker.MockDocker, dryRunDocker *docker.MockDocker,
	normalKubeCtl *kubectl.MockKubeCtl, dryRunKubeCtl *kubectl.MockKubeCtl,
) (*cobra.Command, *options.InitOptions, *BuildOptions, *ApplyOptions, error) {
	rootCommand, initCommand, initInvokerCommands, initOptions, err := setupInitTest(invokers)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	buildCommand, buildOptions := Build(normalDocker, dryRunDocker)
	applyCommand, applyOptions := Apply(normalKubeCtl, dryRunKubeCtl)

	createCommand := Create(initCommand, buildCommand, applyCommand)
	createInvokerCommands := CreateInvokers(invokers, initInvokerCommands, buildCommand, applyCommand)

	rootCommand.AddCommand(createCommand)
	createCommand.AddCommand(createInvokerCommands...)

	return rootCommand, initOptions, buildOptions, applyOptions, err
}
