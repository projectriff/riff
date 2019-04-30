package commands

import (
	"bytes"
	"fmt"

	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("The riff credentials set command", func() {

	var (
		outWriter  bytes.Buffer
		client     core.Client
		clientMock *mocks.Client
		command    *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.Client)
		clientMock = client.(*mocks.Client)
		command = CredentialsSet(&client)
		command.SetOutput(&outWriter)
	})

	AfterEach(func() {
		outWriter.Reset()
		clientMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(command.Use).To(Equal("set"))
		Expect(command.Short).To(Not(BeEmpty()))
		Expect(command.Example).To(Not(BeEmpty()))
	})

	Context("when given wrong args or flags", func() {

		It("should fail if the secret name is not set", func() {
			command.SetArgs([]string{"--namespace", "ns", "--registry", "http://example.com", "--registry-user", "alice"})

			err := command.Execute()

			Expect(err).To(MatchError("flag --secret cannot be empty"))
		})

		It("should fail with any args", func() {
			command.SetArgs([]string{"--namespace", "ns", "unexpected-args"})

			err := command.Execute()

			Expect(err).To(MatchError("accepts 0 arg(s), received 1"))
		})

		It("should fail with an invalid namespace name", func() {
			command.SetArgs([]string{"--secret", "shhh", "--namespace", "inv@l1d!ns", "--gcr", "~/path/to/file", "--docker-hub", "alice"})

			err := command.Execute()

			Expect(err).To(MatchError("when --namespace is set, flag --namespace must be a valid DNS subdomain"))
		})

		It("should fail with conflicting options", func() {
			command.SetArgs([]string{"--secret", "shhh", "--namespace", "ns", "--gcr", "~/path/to/file", "--docker-hub", "alice", "--registry-user", "bob", "--registry", "http://example.com"})

			err := command.Execute()

			Expect(err).To(MatchError("at most one of --gcr, --docker-hub, --registry-user may be set"))
		})

		It("should fail if registry user is set without registry", func() {
			command.SetArgs([]string{"--secret", "shhh", "--namespace", "ns", "--registry-user", "bob"})

			err := command.Execute()

			Expect(err).To(MatchError("when --registry-user is set, flag --registry cannot be empty"))
		})

		It("should fail if registry is set without registry user", func() {
			command.SetArgs([]string{"--secret", "shhh", "--namespace", "ns", "--registry", "http://example.com"})

			err := command.Execute()

			Expect(err).To(MatchError("when --registry is set, flag --registry-user cannot be empty"))
		})

		It("should fail if the registry protocol is not supported", func() {
			command.SetArgs([]string{"--secret", "shhh", "--namespace", "ns", "--registry", "ftp://example.com", "--registry-user", "alice"})

			err := command.Execute()

			Expect(err).To(MatchError("when --registry is set, valid protocols are: \"http\", \"https\", found: \"ftp\""))
		})

		It("should fail if the secret name is blank", func() {
			command.SetArgs([]string{"--secret", "", "--namespace", "ns", "--registry", "http://example.com", "--registry-user", "alice"})

			err := command.Execute()

			Expect(err).To(MatchError("flag --secret cannot be empty"))
		})
	})

	Context("when given suitable args and flags", func() {

		It("involves the client", func() {
			command.SetArgs([]string{"--secret", "s3cr3t", "--namespace", "ns", "--docker-hub", "janedoe"})
			options := core.SetCredentialsOptions{
				SecretName:    "s3cr3t",
				NamespaceName: "ns",
				DockerHubId:   "janedoe",
			}
			clientMock.On("SetCredentials", options).Return(nil)

			err := command.Execute()

			Expect(err).To(BeNil())
			Expect(outWriter.String()).To(HaveSuffix("set completed successfully\n"))
		})

		It("propagates the client errors", func() {
			command.SetArgs([]string{"--secret", "s3cr3t", "--namespace", "ns", "--docker-hub", "janedoe"})
			expectedError := fmt.Errorf("oopsie")
			clientMock.On("SetCredentials", mock.Anything).Return(expectedError)

			err := command.Execute()

			Expect(err).To(MatchError(expectedError))
		})
	})
})
