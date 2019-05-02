package commands

import (
	"bytes"
	"fmt"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"

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

		It("should fail if the secret name is blank", func() {
			command.SetArgs([]string{"--secret", "", "--namespace", "ns", "--registry", "http://example.com", "--registry-user", "alice"})

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

		It("involves the client with the default secret name", func() {
			command.SetArgs([]string{"--namespace", "ns", "--docker-hub", "janedoe"})
			options := core.SetCredentialsOptions{
				SecretName:    "push-credentials",
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

var _ = Describe("The riff credentials list command", func() {

	var (
		outWriter  bytes.Buffer
		client     core.Client
		clientMock *mocks.Client
		command    *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.Client)
		clientMock = client.(*mocks.Client)
		command = CredentialsList(&client)
		command.SetOutput(&outWriter)
	})

	AfterEach(func() {
		outWriter.Reset()
		clientMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(command.Use).To(Equal("list"))
		Expect(command.Short).To(Not(BeEmpty()))
		Expect(command.Example).To(Not(BeEmpty()))
	})

	Context("when given wrong args or flags", func() {

		It("should fail with args", func() {
			command.SetArgs([]string{"something"})

			err := command.Execute()

			Expect(err).To(MatchError("accepts 0 arg(s), received 1"))
		})

		It("should fail with an invalid namespace name", func() {
			command.SetArgs([]string{"--namespace", "inv@l1d!ns"})

			err := command.Execute()

			Expect(err).To(MatchError("when --namespace is set, flag --namespace must be a valid DNS subdomain"))
		})
	})

	const credentialsListOutput = `NAME 
foo  
bar  
baz  
`

	Context("when given suitable args and flags", func() {
		It("should list the credentials of the default namespace", func() {
			command.SetArgs([]string{})
			options := core.ListCredentialsOptions{}
			clientMock.On("ListCredentials", options).Return(&v1.SecretList{
				Items: []v1.Secret{secret("foo"), secret("bar"), secret("baz")},
			}, nil)

			err := command.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(outWriter.String()).To(Equal(credentialsListOutput))
		})

		It("should list the credentials of the specified namespace", func() {
			command.SetArgs([]string{"--namespace", "ns"})
			options := core.ListCredentialsOptions{NamespaceName: "ns"}
			clientMock.On("ListCredentials", options).Return(&v1.SecretList{
				Items: []v1.Secret{secret("foo"), secret("bar"), secret("baz")},
			}, nil)

			err := command.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(outWriter.String()).To(Equal(credentialsListOutput))
		})

		It("should propagate the client error", func() {
			expectedError := fmt.Errorf("oopsie")
			options := core.ListCredentialsOptions{}
			clientMock.On("ListCredentials", options).Return(nil, expectedError)

			err := command.Execute()

			Expect(err).To(MatchError(expectedError))
		})
	})

})

func secret(name string) v1.Secret {
	result := v1.Secret{}
	result.Name = name
	return result
}
