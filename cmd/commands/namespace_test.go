package commands_test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
)

var _ = Describe("The riff namespace destroy create command", func() {
	var (
		outWriter     bytes.Buffer
		kubectlClient core.KubectlClient
		kubectlMock   *mocks.KubectlClient
		command       *cobra.Command
	)

	BeforeEach(func() {
		kubectlClient = new(mocks.KubectlClient)
		kubectlMock = kubectlClient.(*mocks.KubectlClient)
		command = commands.NamespaceCleanup(&kubectlClient)
		command.SetOutput(&outWriter)
	})

	AfterEach(func() {
		outWriter.Reset()
		kubectlMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(command.Use).To(Equal("cleanup"))
		Expect(command.Short).To(Not(BeEmpty()))
		Expect(command.Long).To(Not(BeEmpty()))
		Expect(command.Example).To(Not(BeEmpty()))
	})

	Context("when given wrong args or flags", func() {

		It("should fail with no args", func() {
			command.SetArgs([]string{})

			err := command.Execute()

			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})

		It("should fail with an invalid namespace name", func() {
			command.SetArgs([]string{".invalid-ns"})

			err := command.Execute()

			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})

		It("should fail if asked to remove the default namespace name", func() {
			command.SetArgs([]string{"default", "--remove-ns"})

			err := command.Execute()

			Expect(err).To(MatchError("cleanup canceled: the default namespace cannot be removed"))
		})
	})

	Context("when given suitable args and flags", func() {

		It("should involve the kubectl client with the default options", func() {
			namespace := "ns"
			command.SetArgs([]string{namespace})
			options := core.NamespaceCleanupOptions{NamespaceName: namespace, RemoveNamespace: false}
			kubectlMock.On("NamespaceCleanup", options).Return(nil)

			err := command.Execute()

			Expect(err).To(BeNil())
			s := outWriter.String()
			Expect(s).To(HaveSuffix("cleanup completed successfully\n"))
		})

		It("should involve the kubectl client with the explicit remove-ns option", func() {
			namespace := "ns"
			command.SetArgs([]string{namespace, "--remove-ns"})
			options := core.NamespaceCleanupOptions{NamespaceName: namespace, RemoveNamespace: true}
			kubectlMock.On("NamespaceCleanup", options).Return(nil)

			err := command.Execute()

			Expect(err).To(BeNil())
			s := outWriter.String()
			Expect(s).To(HaveSuffix("cleanup completed successfully\n"))
		})

		It("should propagate kubectl client errors", func() {
			namespace := "ns"
			command.SetArgs([]string{namespace})
			options := core.NamespaceCleanupOptions{NamespaceName: namespace, RemoveNamespace: false}
			expectedError := fmt.Errorf("nope")
			kubectlMock.On("NamespaceCleanup", options).Return(expectedError)

			err := command.Execute()

			Expect(err).To(MatchError(expectedError))
		})
	})
})
