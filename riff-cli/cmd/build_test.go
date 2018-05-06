/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/riff-cli/pkg/docker"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
)

var _ = Describe("The build command", func() {
	var (
		normalDocker *docker.MockDocker
		dryRunDocker *docker.MockDocker
		buildCommand *cobra.Command

		oldCWD string
	)

	BeforeEach(func() {
		var err error
		normalDocker = new(docker.MockDocker)
		dryRunDocker = new(docker.MockDocker)
		buildCommand, _ = Build(normalDocker, dryRunDocker)

		oldCWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		normalDocker.AssertExpectations(GinkgoT())
		dryRunDocker.AssertExpectations(GinkgoT())

		os.Chdir(oldCWD)

	})

	It("should use current working directory as implicit function path", func() {
		os.Chdir("../test_data/node/square")
		buildCommand.SetArgs([]string{"-u", "foo", "-v", "123", "-n", "carre"})

		normalDocker.On("Exec", "build", "-t", "foo/carre:123", ".").Return(nil)

		err := buildCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should accept function path as a flag", func() {
		os.Chdir("../test_data/node")
		buildCommand.SetArgs([]string{"-u", "foo", "-v", "123", "-f", "square", "-n", "carre"})

		normalDocker.On("Exec", "build", "-t", "foo/carre:123", "square").Return(nil)

		err := buildCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should accept function path as an arg", func() {
		os.Chdir("../test_data/node")
		buildCommand.SetArgs([]string{"-u", "foo", "-v", "123", "-n", "carre", "square"})

		normalDocker.On("Exec", "build", "-t", "foo/carre:123", "square").Return(nil)

		err := buildCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should infer function name from path", func() {
		os.Chdir("../test_data/node")
		buildCommand.SetArgs([]string{"-u", "foo", "-v", "123", "square"})

		normalDocker.On("Exec", "build", "-t", "foo/square:123", "square").Return(nil)

		err := buildCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should also push when asked to", func() {
		os.Chdir("../test_data/node")
		buildCommand.SetArgs([]string{"-u", "foo", "-v", "123", "--push", "square"})

		normalDocker.On("Exec", "build", "-t", "foo/square:123", "square").Return(nil)
		normalDocker.On("Exec", "push", "foo/square:123").Return(nil)

		err := buildCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should infer docker username from os username", func() {
		os.Chdir("../test_data/node/square")
		buildCommand.SetArgs([]string{"-v", "123"})
		osuser := osutils.GetCurrentUsername()

		normalDocker.On("Exec", "build", "-t", osuser+"/square:123", ".").Return(nil)

		err := buildCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should require docker username to be lowercase", func() {
		os.Chdir("../test_data/node/square")
		user := "Foo"
		buildCommand.SetArgs([]string{"-u", user})
		err := buildCommand.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(fmt.Sprintf("user account name %s must be lower case", user)))
	})

	It("should require function name to be lowercase", func() {
		os.Chdir("../test_data/node/square")
		name := "squAre"
		buildCommand.SetArgs([]string{"-n", name})
		err := buildCommand.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(fmt.Sprintf("function name %s must be lower case", name)))
	})

	It("should work with no parameters at all", func() {
		os.Chdir("../test_data/node/square")
		buildCommand.SetArgs([]string{})
		osuser := osutils.GetCurrentUsername()

		normalDocker.On("Exec", "build", "-t", osuser+"/square:0.0.1", ".").Return(nil)

		err := buildCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should percolate docker client errors", func() {
		os.Chdir("../test_data/node/square")
		buildCommand.SetArgs([]string{"-u", "foo"})

		normalDocker.On("Exec", "build", "-t", "foo/square:0.0.1", ".").Return(fmt.Errorf("Expected docker error"))

		err := buildCommand.Execute()
		Expect(err).To(MatchError("Expected docker error"))
	})

	It("should not use the real docker client when dry-run is set", func() {
		buildCommand.SetArgs([]string{"--dry-run", "-u", "foo", "-n", "fname", "-v", "123", "--push"})

		dryRunDocker.On("Exec", "build", "-t", "foo/fname:123", ".").Return(nil)
		dryRunDocker.On("Exec", "push", "foo/fname:123").Return(nil)

		buildCommand.Execute()
	})
})
