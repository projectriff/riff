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
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/riff-cli/global"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

var _ = Describe("The version command", func() {
	var (
		kubeClient     *kubectl.MockKubeCtl
		kubeArgs       []string
		writer         bytes.Buffer
		versionCommand *cobra.Command
	)

	BeforeEach(func() {
		global.CLI_VERSION = "0.0.1-testing"
		kubeClient = new(kubectl.MockKubeCtl)
		kubeArgs = []string{
			"get", "deployments",
			"--all-namespaces",
			"-l", "app=riff",
			"-o=custom-columns=COMPONENT:.metadata.labels.component,IMAGE:.spec.template.spec.containers[0].image",
		}
		writer = bytes.Buffer{}
		versionCommand = Version(&writer, kubeClient)
	})

	AfterEach(func() {
		kubeClient.AssertExpectations(GinkgoT())
	})

	It("should list the cli and component versions", func() {
		kubeClient.On("Exec", kubeArgs).Return("<Component Table>", nil)

		err := versionCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(writer.String()).To(Equal(fmt.Sprintf("riff CLI version: %s\n\n%s", "0.0.1-testing", "<Component Table>")))
	})

	It("should list the cli version even when kubectl fails", func() {
		kubeClient.On("Exec", kubeArgs).Return("", fmt.Errorf("kubectl fault"))

		err := versionCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(writer.String()).To(Equal(fmt.Sprintf("riff CLI version: %s\n\n%s", "0.0.1-testing", "Unable to list component versions")))
	})

})
