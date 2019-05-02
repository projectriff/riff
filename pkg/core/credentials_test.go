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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

var _ = Describe("credentials", func() {

	var (
		client              core.Client
		kubeClient          *vendor_mocks.Interface
		kubeCtl             *mockkubectl.KubeCtl
		mockCore            *vendor_mocks.CoreV1Interface
		mockSecrets         *vendor_mocks.SecretInterface
		mockServiceAccounts *vendor_mocks.ServiceAccountInterface
		mockClientConfig    *vendor_mocks.ClientConfig
	)

	BeforeEach(func() {
		kubeClient = new(vendor_mocks.Interface)
		kubeCtl = new(mockkubectl.KubeCtl)
		mockCore = new(vendor_mocks.CoreV1Interface)
		mockSecrets = new(vendor_mocks.SecretInterface)
		mockServiceAccounts = new(vendor_mocks.ServiceAccountInterface)
		mockClientConfig = new(vendor_mocks.ClientConfig)
		kubeClient.On("CoreV1").Return(mockCore)
		mockCore.On("ServiceAccounts", mock.Anything).Return(mockServiceAccounts)
		mockCore.On("Secrets", mock.Anything).Return(mockSecrets)

		client = core.NewClient(mockClientConfig, kubeClient, nil, nil, kubeCtl, new(mockkustomize.Kustomizer))
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

		It("successfully updates and binds the secret to the newly created service account", func() {
			secret := secret(secretName)
			mockSecrets.On("Get", secretName, mock.Anything).Return(&secret, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(nil)
			mockSecrets.On("Create", mock.Anything).Return(&secret, nil)
			serviceAccount := serviceAccount(core.BuildServiceAccountName)
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create",
				mock.MatchedBy(andPredicates(
					named(core.BuildServiceAccountName),
					withLabels(map[string]string{"projectriff.io/installer": env.Cli.Name, "projectriff.io/version": env.Cli.Version}),
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

	Describe("ListCredentials", func() {

		BeforeEach(func() {
			mockClientConfig.On("Namespace").Return("default", false, nil)
		})

		It("propagates the underlying client error", func() {
			expectedError := fmt.Errorf("oopsie")
			mockSecrets.On("List", mock.Anything).Return(nil, expectedError)

			_, err := client.ListCredentials(core.ListCredentialsOptions{})

			Expect(err).To(MatchError(expectedError))
		})

		It("returns the matching secrets", func() {
			secrets := &v1.SecretList{}
			mockSecrets.On("List", metav1.ListOptions{LabelSelector: "projectriff.io/installer,projectriff.io/version"}).Return(secrets, nil)

			result, err := client.ListCredentials(core.ListCredentialsOptions{})

			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(Equal(secrets))
		})
	})

	Describe("DeleteCredentials", func() {

		BeforeEach(func() {
			mockClientConfig.On("Namespace").Return("default", false, nil)
		})

		It("propagates the underlying client error", func() {
			expectedError := fmt.Errorf("oopsie")
			mockSecrets.On("Delete", "secret", mock.Anything).Return(expectedError)

			err := client.DeleteCredentials(core.DeleteCredentialsOptions{Name: "secret"})

			Expect(err).To(MatchError(expectedError))
		})

		It("returns the underlying client result", func() {
			mockSecrets.On("Delete", "secret", mock.Anything).Return(nil)

			err := client.DeleteCredentials(core.DeleteCredentialsOptions{Name: "secret"})

			Expect(err).To(Not(HaveOccurred()))
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

func withLabels(labels map[string]string) func(*v1.ServiceAccount) bool {
	return func(account *v1.ServiceAccount) bool {
		return reflect.DeepEqual(labels, account.Labels)
	}
}
