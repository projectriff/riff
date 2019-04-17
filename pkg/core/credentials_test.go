package core_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	mockkustomize "github.com/projectriff/riff/pkg/core/kustomize/mocks"
	"github.com/projectriff/riff/pkg/core/vendor_mocks"
	"github.com/projectriff/riff/pkg/env"
	mockkubectl "github.com/projectriff/riff/pkg/kubectl/mocks"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("credentials", func() {

	var (
		client              core.Client
		kubeClient          *vendor_mocks.Interface
		kubeCtl             *mockkubectl.KubeCtl
		mockCore            *vendor_mocks.CoreV1Interface
		mockSecrets         *vendor_mocks.SecretInterface
		mockServiceAccounts *vendor_mocks.ServiceAccountInterface
	)

	JustBeforeEach(func() {
		kubeClient = new(vendor_mocks.Interface)
		kubeCtl = new(mockkubectl.KubeCtl)
		mockCore = new(vendor_mocks.CoreV1Interface)
		mockSecrets = new(vendor_mocks.SecretInterface)
		mockServiceAccounts = new(vendor_mocks.ServiceAccountInterface)
		kubeClient.On("CoreV1").Return(mockCore)
		mockCore.On("ServiceAccounts", mock.Anything).Return(mockServiceAccounts)
		mockCore.On("Secrets", mock.Anything).Return(mockSecrets)

		client = core.NewClient(nil, kubeClient, nil, nil, kubeCtl, new(mockkustomize.Kustomizer))
	})

	AfterEach(func() {
		mockSecrets.AssertExpectations(GinkgoT())
		mockServiceAccounts.AssertExpectations(GinkgoT())
	})

	Describe("SetCredentials", func() {

		const secretName = "s#cr#t"

		It("fails if the secret check fails", func() {
			expectedError := fmt.Errorf("oopsie")
			mockSecrets.On("Get", secretName, mock.Anything).Return(&v1.Secret{}, expectedError)

			err := client.SetCredentials(core.SetCredentialsOptions{
				NamespaceName: "ns",
				SecretName:    secretName,
				DockerHubId:   "janedoe",
			})

			Expect(err).To(MatchError(expectedError))
		})

		It("fails if the service account does not exist", func() {
			mockSecrets.On("Get", secretName, mock.Anything).Return(nil, notFound())
			secret := secret(secretName)
			mockSecrets.On("Create", mock.MatchedBy(secretNamed(secretName))).Return(&secret, nil)
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())

			err := client.SetCredentials(core.SetCredentialsOptions{
				NamespaceName: "ns",
				SecretName:    secretName,
				DockerHubId:   "janedoe",
			})

			Expect(err).To(MatchError(env.Cli.Name + ` service account not found. Please run "` + env.Cli.Name + ` namespace init" first`))
		})

		It("fails if the service account check fails", func() {
			expectedError := fmt.Errorf("oopsie")
			mockSecrets.On("Get", secretName, mock.Anything).Return(nil, notFound())
			secret := secret(secretName)
			mockSecrets.On("Create", mock.MatchedBy(secretNamed(secretName))).Return(&secret, nil)
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, expectedError)

			err := client.SetCredentials(core.SetCredentialsOptions{
				NamespaceName: "ns",
				SecretName:    secretName,
				DockerHubId:   "janedoe",
			})

			Expect(err).To(MatchError(expectedError))
		})

		It("fails if the secret creation fails", func() {
			expectedError := fmt.Errorf("oopsie")
			mockSecrets.On("Get", secretName, mock.Anything).Return(nil, notFound())
			mockSecrets.On("Create", mock.Anything).Return(nil, expectedError)

			err := client.SetCredentials(core.SetCredentialsOptions{
				NamespaceName: "ns",
				SecretName:    secretName,
				DockerHubId:   "janedoe",
			})

			Expect(err).To(MatchError(expectedError))
		})

		It("fails if the secret binding fails", func() {
			expectedError := fmt.Errorf("oopsie")
			mockSecrets.On("Get", secretName, mock.Anything).Return(nil, notFound())
			secret := secret(secretName)
			mockSecrets.On("Create", mock.Anything).Return(&secret, nil)
			serviceAccount := serviceAccount(core.BuildServiceAccountName)
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(&serviceAccount, nil)
			mockServiceAccounts.On("Update", mock.MatchedBy(named(core.BuildServiceAccountName))).Return(nil, expectedError)

			err := client.SetCredentials(core.SetCredentialsOptions{
				NamespaceName: "ns",
				SecretName:    secretName,
				DockerHubId:   "janedoe",
			})

			Expect(err).To(MatchError(expectedError))
		})

		It("successfully creates and binds the secret to the build service account", func() {
			mockSecrets.On("Get", secretName, mock.Anything).Return(nil, notFound())
			secret := secret(secretName)
			mockSecrets.On("Create", mock.Anything).Return(&secret, nil)
			serviceAccount := serviceAccount(core.BuildServiceAccountName)
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(&serviceAccount, nil)
			mockServiceAccounts.On("Update",
				mock.MatchedBy(andPredicates(
					named(core.BuildServiceAccountName),
					withSingleSecret(secretName),
				))).Return(&serviceAccount, nil)

			err := client.SetCredentials(core.SetCredentialsOptions{
				NamespaceName: "ns",
				SecretName:    secretName,
				GcrTokenPath:  "fixtures/gcr-creds",
			})

			Expect(err).NotTo(HaveOccurred())
		})

		It("fails to update the secret if the deletion fails", func() {
			expectedError := fmt.Errorf("oopsie")
			secret := secret(secretName)
			mockSecrets.On("Get", secretName, mock.Anything).Return(&secret, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(expectedError)

			err := client.SetCredentials(core.SetCredentialsOptions{
				NamespaceName: "ns",
				SecretName:    secretName,
				GcrTokenPath:  "fixtures/gcr-creds",
			})

			Expect(err).To(MatchError(expectedError))
		})

		It("successfully updates and binds the secret to the build service account", func() {
			secret := secret(secretName)
			mockSecrets.On("Get", secretName, mock.Anything).Return(&secret, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(nil)
			mockSecrets.On("Create", mock.Anything).Return(&secret, nil)
			serviceAccount := serviceAccount(core.BuildServiceAccountName)
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(&serviceAccount, nil)
			mockServiceAccounts.On("Update",
				mock.MatchedBy(andPredicates(
					named(core.BuildServiceAccountName),
					withSingleSecret(secretName),
				))).Return(&serviceAccount, nil)

			err := client.SetCredentials(core.SetCredentialsOptions{
				NamespaceName: "ns",
				SecretName:    secretName,
				GcrTokenPath:  "fixtures/gcr-creds",
			})

			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func andPredicates(predicates ...func(*v1.ServiceAccount) bool) func(*v1.ServiceAccount) bool {
	return func(sa *v1.ServiceAccount) bool {
		for _, predicate := range predicates {
			if !predicate(sa) {
				return false
			}
		}
		return true
	}
}

func withSingleSecret(secretName string) func(*v1.ServiceAccount) bool {
	return func(sa *v1.ServiceAccount) bool {
		count := 0
		for _, secret := range sa.Secrets {
			if secret.Name == secretName {
				count++
			}
		}
		return count == 1
	}
}

func secretNamed(secretName string) func(*v1.Secret) bool {
	return func(secret *v1.Secret) bool {
		return secret.Name == secretName
	}
}
