package commands_test

import (
	"bytes"
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
)

var _ = Describe("The riff namespace init command", func() {

	var (
		manifests     map[string]*core.Manifest
		client        core.Client
		clientMock    *mocks.Client
		namespaceInit *cobra.Command
	)

	BeforeEach(func() {
		manifests = map[string]*core.Manifest{}
		client = new(mocks.Client)
		clientMock = client.(*mocks.Client)
		namespaceInit = commands.NamespaceInit(manifests, &client)
		namespaceInit.SetOutput(ioutil.Discard)
	})

	Context("when given wrong args or flags", func() {

		It("fails with no args", func() {
			namespaceInit.SetArgs([]string{})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})

		It("fails with invalid namespace name", func() {
			namespaceInit.SetArgs([]string{".invalid"})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})

		DescribeTable("fails with too many auth configuration modes",
			func(modes ...[]string) {
				namespaceInit.SetArgs(concat([]string{"ns", "--manifest", "some-path"}, concat(modes...)))

				err := namespaceInit.Execute()

				Expect(err).To(MatchError("at most one of --gcr, --dockerhub, --no-secret, --registry-host must be set"))
			},
			Entry("all modes					=>", []string{"--dockerhub", "/path/to/.docker/config.json"}, []string{"--gcr", "/path/to/gcr.json"}, []string{"--no-secret"}, []string{"--registry-host", "registry.example.com"}),
			Entry("docker+grc+nosecret		=>", []string{"--dockerhub", "/path/to/.docker/config.json"}, []string{"--gcr", "/path/to/gcr.json"}, []string{"--no-secret"}),
			Entry("docker+grc+registry		=>", []string{"--dockerhub", "/path/to/.docker/config.json"}, []string{"--gcr", "/path/to/gcr.json"}, []string{"--registry-host", "registry.example.com"}),
			Entry("docker+grc				=>", []string{"--dockerhub", "/path/to/.docker/config.json"}, []string{"--gcr", "/path/to/gcr.json"}),
			Entry("docker+nosecret+registry	=>", []string{"--dockerhub", "/path/to/.docker/config.json"}, []string{"--no-secret"}, []string{"--registry-host", "registry.example.com"}),
			Entry("docker+nosecret			=>", []string{"--dockerhub", "/path/to/.docker/config.json"}, []string{"--no-secret"}),
			Entry("docker+registry			=>", []string{"--dockerhub", "/path/to/.docker/config.json"}, []string{"--registry-host", "registry.example.com"}),
			Entry("gcr+nosecret+registry		=>", []string{"--gcr", "/path/to/gcr.json"}, []string{"--no-secret"}, []string{"--registry-host", "registry.example.com"}),
			Entry("gcr+nosecret 				=>", []string{"--gcr", "/path/to/gcr.json"}, []string{"--no-secret"}),
			Entry("gcr+registry 				=>", []string{"--gcr", "/path/to/gcr.json"}, []string{"--registry-host", "registry.example.com"}),
			Entry("nosecret+registry			=>", []string{"--no-secret"}, []string{"--registry-host", "registry.example.com"}),
		)

		It("fails with ambiguous secret configuration", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--secret", "my-secret", "--no-secret"})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError("at most one of --secret, --no-secret must be set"))
		})

		It("fails with blank secret name", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--secret", ""})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError("flag --secret cannot be empty"))
		})

		It("fails with no pre-existing manifest and no explicit manifest path", func() {
			namespaceInit.SetArgs([]string{"ns", "--no-secret"})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError("required flag(s) \"manifest\" not set"))
		})

		It("fails if the registry server is set but not the registry username", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--registry-host", "registry.example.com"})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError("when --registry-host is set, flag --registry-user cannot be empty"))
		})

		It("fails if the registry user is set but not the registry host", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--registry-user", "me"})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError("when --registry-user is set, flag --registry-host cannot be empty"))
		})

		It("fails if the registry protocol is set but not the registry host", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--registry-protocol", "http"})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError("when --registry-protocol is set, flag --registry-host cannot be empty"))
		})

		It("fails if the registry protocol is not supported", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--registry-protocol", "ftp", "--registry-host", "registry.example.com", "--registry-user", "me"})

			err := namespaceInit.Execute()

			Expect(err).To(MatchError(`when --registry-host is set, flag --registry-protocol cannot have value "ftp", valid values are: "https", "http"`))
		})
	})

	Context("when given suitable args and flags", func() {

		It("involves the core.Client", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--secret", "s3cr3t"})
			clientMock.On("NamespaceInit", manifests, core.NamespaceInitOptions{
				NamespaceName:    "ns",
				Manifest:         "some-path",
				SecretName:       "s3cr3t",
				RegistryProtocol: "https",
			}).Return(nil)

			err := namespaceInit.Execute()

			Expect(err).NotTo(HaveOccurred())
		})

		It("involves the core.Client with GCR config", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--gcr", "/path/to/gcr/config.json"})
			clientMock.On("NamespaceInit", manifests, core.NamespaceInitOptions{
				NamespaceName:    "ns",
				Manifest:         "some-path",
				GcrTokenPath:     "/path/to/gcr/config.json",
				SecretName:       "push-credentials",
				RegistryProtocol: "https",
			}).Return(nil)

			err := namespaceInit.Execute()

			Expect(err).NotTo(HaveOccurred())
		})

		It("involves the core.Client with Dockerhub config", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--dockerhub", "username"})
			clientMock.On("NamespaceInit", manifests, core.NamespaceInitOptions{
				NamespaceName:     "ns",
				Manifest:          "some-path",
				DockerHubUsername: "username",
				SecretName:        "push-credentials",
				RegistryProtocol:  "https",
			}).Return(nil)

			err := namespaceInit.Execute()

			Expect(err).NotTo(HaveOccurred())
		})

		It("involves the core.Client without any secret", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--no-secret"})
			clientMock.On("NamespaceInit", manifests, core.NamespaceInitOptions{
				NamespaceName:    "ns",
				Manifest:         "some-path",
				NoSecret:         true,
				SecretName:       "push-credentials",
				RegistryProtocol: "https",
			}).Return(nil)

			err := namespaceInit.Execute()

			Expect(err).NotTo(HaveOccurred())
		})

		It("involves the core.Client with explicit registry configuration", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--registry-host", "registry.example.com", "--registry-user", "me"})
			clientMock.On("NamespaceInit", manifests, core.NamespaceInitOptions{
				NamespaceName:    "ns",
				Manifest:         "some-path",
				RegistryProtocol: "https",
				RegistryHost:     "registry.example.com",
				RegistryUser:     "me",
				SecretName:       "push-credentials",
			}).Return(nil)

			err := namespaceInit.Execute()

			Expect(err).NotTo(HaveOccurred())
		})

		It("propagates the core.Client errors", func() {
			namespaceInit.SetArgs([]string{"ns", "--manifest", "some-path", "--secret", "s3cr3t"})
			expectedError := fmt.Errorf("oopsie")
			clientMock.On("NamespaceInit", manifests, core.NamespaceInitOptions{
				NamespaceName:    "ns",
				Manifest:         "some-path",
				SecretName:       "s3cr3t",
				RegistryProtocol: "https",
			}).Return(expectedError)

			err := namespaceInit.Execute()

			Expect(err).To(MatchError(expectedError))
		})
	})
})

var _ = Describe("The riff namespace cleanup command", func() {
	var (
		outWriter  bytes.Buffer
		client     core.Client
		clientMock *mocks.Client
		command    *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.Client)
		clientMock = client.(*mocks.Client)
		command = commands.NamespaceCleanup(&client)
		command.SetOutput(&outWriter)
	})

	AfterEach(func() {
		outWriter.Reset()
		clientMock.AssertExpectations(GinkgoT())
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
			clientMock.On("NamespaceCleanup", options).Return(nil)

			err := command.Execute()

			Expect(err).To(BeNil())
			s := outWriter.String()
			Expect(s).To(HaveSuffix("cleanup completed successfully\n"))
		})

		It("should involve the kubectl client with the explicit remove-ns option", func() {
			namespace := "ns"
			command.SetArgs([]string{namespace, "--remove-ns"})
			options := core.NamespaceCleanupOptions{NamespaceName: namespace, RemoveNamespace: true}
			clientMock.On("NamespaceCleanup", options).Return(nil)

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
			clientMock.On("NamespaceCleanup", options).Return(expectedError)

			err := command.Execute()

			Expect(err).To(MatchError(expectedError))
		})
	})
})

func concat(arrays ...[]string) []string {
	var result []string
	switch len(arrays) {
	case 0:
		return result
	default:
		for _, array := range arrays {
			result = append(result, array...)
		}
		return result
	}
}
