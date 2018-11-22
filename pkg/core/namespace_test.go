/*
 * Copyright 2018 The original author or authors
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

package core

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/projectriff/riff/pkg/core/vendor_mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("The NamespaceInit function", func() {

	var (
		kubectlClient       KubectlClient
		kubeClient          *vendor_mocks.Interface
		mockCore            *vendor_mocks.CoreV1Interface
		mockNamespaces      *vendor_mocks.NamespaceInterface
		mockServiceAccounts *vendor_mocks.ServiceAccountInterface
		mockSecrets         *vendor_mocks.SecretInterface
		manifests           map[string]*Manifest
	)

	JustBeforeEach(func() {
		kubeClient = new(vendor_mocks.Interface)
		mockCore = new(vendor_mocks.CoreV1Interface)
		mockNamespaces = new(vendor_mocks.NamespaceInterface)
		mockServiceAccounts = new(vendor_mocks.ServiceAccountInterface)
		mockSecrets = new(vendor_mocks.SecretInterface)

		kubeClient.On("CoreV1").Return(mockCore)
		mockCore.On("Namespaces").Return(mockNamespaces)
		mockCore.On("ServiceAccounts", mock.Anything).Return(mockServiceAccounts)
		mockCore.On("Secrets", mock.Anything).Return(mockSecrets)

		kubectlClient = NewKubectlClient(kubeClient)
	})

	AfterEach(func() {
		mockNamespaces.AssertExpectations(GinkgoT())
		mockServiceAccounts.AssertExpectations(GinkgoT())
		mockSecrets.AssertExpectations(GinkgoT())
	})

	It("should fail on wrong manifest", func() {
		options := NamespaceInitOptions{Manifest: "wrong"}
		err := kubectlClient.NamespaceInit(manifests, options)
		Expect(err).To(MatchError(ContainSubstring("wrong: "))) // error message is quite different on Windows and macOS
	})

	It("should create namespace and sa if needed", func() {

		options := NamespaceInitOptions{
			Manifest:      "fixtures/empty.yaml",
			NamespaceName: "foo",
			SecretName:    "push-credentials",
		}

		namespace := &v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: "foo"}}
		mockNamespaces.On("Get", "foo", mock.Anything).Return(nil, notFound())
		mockNamespaces.On("Create", namespace).Return(namespace, nil)

		mockSecrets.On("Get", "push-credentials", meta_v1.GetOptions{}).Return(&v1.Secret{}, nil)

		serviceAccount := &v1.ServiceAccount{}
		mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(nil, notFound())
		mockServiceAccounts.On("Create", mock.MatchedBy(named(serviceAccountName))).Return(serviceAccount, nil)

		err := kubectlClient.NamespaceInit(manifests, options)
		Expect(err).To(Not(HaveOccurred()))
	})

	It("should create secret for gcr", func() {

		options := NamespaceInitOptions{
			Manifest:      "fixtures/empty.yaml",
			NamespaceName: "foo",
			GcrTokenPath:  "fixtures/gcr-creds",
			SecretName:    "push-credentials",
		}

		namespace := &v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: "foo"}}
		mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

		serviceAccount := &v1.ServiceAccount{}
		mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(serviceAccount, nil)

		secret := &v1.Secret{}
		mockSecrets.On("Delete", "push-credentials", &meta_v1.DeleteOptions{}).Return(nil)
		mockSecrets.On("Create", mock.Anything).Run(func(args mock.Arguments) {
			s := args[0].(*v1.Secret)
			Expect(s.StringData).To(HaveKeyWithValue("username", "_json_key"))
			Expect(s.StringData).To(HaveKeyWithValue("password", "hush hush"))
		}).Return(secret, nil)

		mockServiceAccounts.On("Update", mock.Anything).Run(func(args mock.Arguments) {
			sa := args[0].(*v1.ServiceAccount)
			Expect(sa.Secrets).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Name": Equal("push-credentials"),
			})))
		}).Return(serviceAccount, nil)

		err := kubectlClient.NamespaceInit(manifests, options)
		Expect(err).To(Not(HaveOccurred()))
	})

	Context("when dealing with Dockerhub", func() {

		var oldStdIn *os.File

		BeforeEach(func() {
			oldStdIn = os.Stdin
			creds, _ := os.Open("fixtures/dockerhub-creds")
			os.Stdin = creds
		})

		AfterEach(func() {
			os.Stdin = oldStdIn
		})

		It("should create secret for dockerhub", func() {

			options := NamespaceInitOptions{
				Manifest:          "fixtures/empty.yaml",
				NamespaceName:     "foo",
				DockerHubUsername: "roger",
				SecretName:        "push-credentials",
			}

			namespace := &v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(serviceAccount, nil)

			secret := &v1.Secret{}
			mockSecrets.On("Delete", "push-credentials", &meta_v1.DeleteOptions{}).Return(nil)
			mockSecrets.On("Create", mock.Anything).Run(func(args mock.Arguments) {
				s := args[0].(*v1.Secret)
				Expect(s.StringData).To(HaveKeyWithValue("username", "roger"))
				Expect(s.StringData).To(HaveKeyWithValue("password", "s3cr3t"))
			}).Return(secret, nil)

			mockServiceAccounts.On("Update", mock.Anything).Run(func(args mock.Arguments) {
				sa := args[0].(*v1.ServiceAccount)
				Expect(sa.Secrets).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("push-credentials"),
				})))
			}).Return(serviceAccount, nil)

			err := kubectlClient.NamespaceInit(manifests, options)
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	It("should run unauthenticated and still create a service account", func() {
		options := NamespaceInitOptions{
			Manifest:      "fixtures/empty.yaml",
			NamespaceName: "foo",
			NoSecret:      true,
		}

		namespace := &v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: "foo"}}
		mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

		serviceAccount := &v1.ServiceAccount{}
		mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(nil, notFound())
		mockServiceAccounts.On("Create", mock.MatchedBy(named(serviceAccountName))).Return(serviceAccount, nil)

		err := kubectlClient.NamespaceInit(manifests, options)
		Expect(err).To(Not(HaveOccurred()))
	})

})

func notFound() *errors.StatusError {
	return errors.NewNotFound(schema.GroupResource{}, "")
}

func named(name string) func(sa *v1.ServiceAccount) bool {
	return func(sa *v1.ServiceAccount) bool {
		return sa.Name == name
	}
}
