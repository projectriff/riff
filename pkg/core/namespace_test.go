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
	"fmt"
	"github.com/projectriff/riff/pkg/core/kustomize"
	"github.com/projectriff/riff/pkg/core/kustomize/mocks"
	"github.com/projectriff/riff/pkg/env"
	"github.com/projectriff/riff/pkg/kubectl"
	"github.com/projectriff/riff/pkg/kubectl/mocks"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"net/url"
	"os"
	"path/filepath"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/projectriff/riff/pkg/core/vendor_mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Namespace-related functions, such as", func() {

	var (
		kubectlClient              KubectlClient
		kubeClient                 *vendor_mocks.Interface
		kubeCtl                    *mockkubectl.KubeCtl
		mockCore                   *vendor_mocks.CoreV1Interface
		mockNamespaces             *vendor_mocks.NamespaceInterface
		mockServiceAccounts        *vendor_mocks.ServiceAccountInterface
		mockSecrets                *vendor_mocks.SecretInterface
		mockPersistentVolumeClaims *vendor_mocks.PersistentVolumeClaimInterface
		mockKustomizer             *mockkustomize.Kustomizer
		manifests                  map[string]*Manifest
	)

	JustBeforeEach(func() {
		kubeClient = new(vendor_mocks.Interface)
		kubeCtl = new(mockkubectl.KubeCtl)
		mockCore = new(vendor_mocks.CoreV1Interface)
		mockNamespaces = new(vendor_mocks.NamespaceInterface)
		mockServiceAccounts = new(vendor_mocks.ServiceAccountInterface)
		mockSecrets = new(vendor_mocks.SecretInterface)
		mockPersistentVolumeClaims = new(vendor_mocks.PersistentVolumeClaimInterface)
		mockKustomizer = new(mockkustomize.Kustomizer)
		manifests = map[string]*Manifest{}

		kubeClient.On("CoreV1").Return(mockCore)
		mockCore.On("Namespaces").Return(mockNamespaces)
		mockCore.On("ServiceAccounts", mock.Anything).Return(mockServiceAccounts)
		mockCore.On("Secrets", mock.Anything).Return(mockSecrets)
		mockCore.On("PersistentVolumeClaims", mock.Anything).Return(mockPersistentVolumeClaims)

		kubectlClient = makeKubectlClient(kubeClient, kubeCtl, mockKustomizer)
	})

	AfterEach(func() {
		mockNamespaces.AssertExpectations(GinkgoT())
		mockServiceAccounts.AssertExpectations(GinkgoT())
		mockSecrets.AssertExpectations(GinkgoT())
		mockPersistentVolumeClaims.AssertExpectations(GinkgoT())
	})

	Describe("the NamespaceCleanup function", func() {

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
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(nil, notFound())
			mockNamespaces.On("Create", namespace).Return(namespace, nil)
			mockSecrets.On("Get", "push-credentials", metav1.GetOptions{}).Return(&v1.Secret{}, nil)
			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(nil, notFound())
			labels := map[string]string{"created-by": env.Cli.Name + "-" + env.Cli.Version}
			mockServiceAccounts.On("Create", mock.MatchedBy(namedAndLabelled(serviceAccountName, labels))).Return(serviceAccount, nil)

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

			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(serviceAccount, nil)

			secret := &v1.Secret{}
			mockSecrets.On("Delete", "push-credentials", &metav1.DeleteOptions{}).Return(nil)
			mockSecrets.On("Create", mock.Anything).Run(func(args mock.Arguments) {
				s := args[0].(*v1.Secret)
				Expect(s.StringData).To(HaveKeyWithValue("username", "_json_key"))
				Expect(s.StringData).To(HaveKeyWithValue("password", "hush hush"))
				Expect(s.Labels["created-by"]).To(HavePrefix(env.Cli.Name))
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

				namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
				mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

				serviceAccount := &v1.ServiceAccount{}
				mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(serviceAccount, nil)

				secret := &v1.Secret{}
				mockSecrets.On("Delete", "push-credentials", &metav1.DeleteOptions{}).Return(nil)
				mockSecrets.On("Create", mock.Anything).Run(func(args mock.Arguments) {
					s := args[0].(*v1.Secret)
					Expect(s.StringData).To(HaveKeyWithValue("username", "roger"))
					Expect(s.StringData).To(HaveKeyWithValue("password", "s3cr3t"))
					Expect(s.Labels["created-by"]).To(HavePrefix(env.Cli.Name))
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

			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(serviceAccountName))).Return(serviceAccount, nil)

			err := kubectlClient.NamespaceInit(manifests, options)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should apply label to namespace resource", func() {
			options := NamespaceInitOptions{
				Manifest:      "stable",
				NamespaceName: "foo",
				NoSecret:      true,
			}
			namespaceResource := unsafeAbs("fixtures/initial_pvc.yaml")
			manifests["stable"] = &Manifest{
				Namespace: []string{namespaceResource},
			}
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(serviceAccountName))).Return(serviceAccount, nil)
			customizedResourceContents := contentsOf("fixtures/kustom_pvc.yaml")
			mockKustomizer.On("ApplyLabels",
				mock.MatchedBy(urlPath(namespaceResource)),
				mock.MatchedBy(key("created-by"))).Return(customizedResourceContents, nil)
			kubeCtl.On("ExecStdin", []string{"apply", "-n", "foo", "-f", "-"}, &customizedResourceContents).
				Return("done!", nil)

			err := kubectlClient.NamespaceInit(manifests, options)

			Expect(err).To(Not(HaveOccurred()))
		})

		It("should fail if the PVC label kustomization fails", func() {
			options := NamespaceInitOptions{
				Manifest:      "stable",
				NamespaceName: "foo",
				NoSecret:      true,
			}
			namespaceResource := unsafeAbs("fixtures/initial_pvc.yaml")
			manifests["stable"] = &Manifest{
				Namespace: []string{namespaceResource},
			}
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", serviceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(serviceAccountName))).Return(serviceAccount, nil)
			expectedError := fmt.Errorf("kustomization failed")
			mockKustomizer.On("ApplyLabels",
				mock.MatchedBy(urlPath(namespaceResource)),
				mock.MatchedBy(key("created-by"))).Return(nil, expectedError)

			err := kubectlClient.NamespaceInit(manifests, options)

			Expect(err).To(MatchError(expectedError))
		})
	})

	Describe("the NamespaceCleanup function", func() {

		var (
			namespace           string
			options             NamespaceCleanupOptions
			expectedListOptions metav1.ListOptions
		)

		BeforeEach(func() {
			namespace = "foo"
			options = NamespaceCleanupOptions{
				NamespaceName: namespace,
			}
			expectedListOptions = metav1.ListOptions{LabelSelector: "created-by=" + env.Cli.Name + "-" + env.Cli.Version}
		})

		It("should fail if the service account list fails", func() {
			expectedError := fmt.Errorf("SA list failed")
			mockServiceAccounts.On("List", expectedListOptions).Return(nil, expectedError)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should fail if the service account deletion fails", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(serviceAccountName)},
			}, nil)
			expectedError := fmt.Errorf("SA deletion failed")
			mockServiceAccounts.On("Delete", serviceAccountName, mock.Anything).Return(expectedError)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(MatchError(fmt.Sprintf("Unable to delete service account %s: %s", serviceAccountName, expectedError.Error())))
		})

		It("should fail if the persistent volume claim list fails", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(serviceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", serviceAccountName, mock.Anything).Return(nil)
			expectedError := fmt.Errorf("PVC list failed")
			mockPersistentVolumeClaims.On("List", expectedListOptions).Return(nil, expectedError)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should fail if the persistent volume claim deletion fails", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(serviceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", serviceAccountName, mock.Anything).Return(nil)
			pvcName := "pvc-name"
			mockPersistentVolumeClaims.On("List", expectedListOptions).Return(&v1.PersistentVolumeClaimList{
				Items: []v1.PersistentVolumeClaim{persistentVolumeClaim(pvcName)},
			}, nil)
			expectedError := fmt.Errorf("PVC deletion failed")
			mockPersistentVolumeClaims.On("Delete", pvcName, mock.Anything).Return(expectedError)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(MatchError(fmt.Sprintf("Unable to delete persistent volume claim %s: %s", pvcName, expectedError.Error())))
		})

		It("should fail if the secret list fails", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(serviceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", serviceAccountName, mock.Anything).Return(nil)
			pvcName := "pvc-name"
			mockPersistentVolumeClaims.On("List", expectedListOptions).Return(&v1.PersistentVolumeClaimList{
				Items: []v1.PersistentVolumeClaim{persistentVolumeClaim(pvcName)},
			}, nil)
			mockPersistentVolumeClaims.On("Delete", pvcName, mock.Anything).Return(nil)
			expectedError := fmt.Errorf("secret deletion failed")
			mockSecrets.On("List", expectedListOptions).Return(nil, expectedError)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should fail if the secret deletion fails", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(serviceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", serviceAccountName, mock.Anything).Return(nil)
			pvcName := "pvc-name"
			mockPersistentVolumeClaims.On("List", expectedListOptions).Return(&v1.PersistentVolumeClaimList{
				Items: []v1.PersistentVolumeClaim{persistentVolumeClaim(pvcName)},
			}, nil)
			mockPersistentVolumeClaims.On("Delete", pvcName, mock.Anything).Return(nil)
			secretName := "s3cr3t"
			mockSecrets.On("List", expectedListOptions).Return(&v1.SecretList{
				Items: []v1.Secret{secret(secretName)},
			}, nil)
			expectedError := fmt.Errorf("secret deletion failed")
			mockSecrets.On("Delete", secretName, mock.Anything).Return(expectedError)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(MatchError(fmt.Sprintf("Unable to delete secret %s: %s", secretName, expectedError.Error())))
		})

		It("should fail if the namespace deletion fails", func() {
			options.RemoveNamespace = true
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(serviceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", serviceAccountName, mock.Anything).Return(nil)
			pvcName := "pvc-name"
			mockPersistentVolumeClaims.On("List", expectedListOptions).Return(&v1.PersistentVolumeClaimList{
				Items: []v1.PersistentVolumeClaim{persistentVolumeClaim(pvcName)},
			}, nil)
			mockPersistentVolumeClaims.On("Delete", pvcName, mock.Anything).Return(nil)
			secretName := "s3cr3t"
			mockSecrets.On("List", expectedListOptions).Return(&v1.SecretList{
				Items: []v1.Secret{secret(secretName)},
			}, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(nil)
			expectedError := fmt.Errorf("namespace deletion failed")
			mockNamespaces.On("Delete", namespace, mock.Anything).Return(expectedError)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should successfully delete the service account, the PVC and the secret", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(serviceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", serviceAccountName, mock.Anything).Return(nil)
			pvcName := "pvc-name"
			mockPersistentVolumeClaims.On("List", expectedListOptions).Return(&v1.PersistentVolumeClaimList{
				Items: []v1.PersistentVolumeClaim{persistentVolumeClaim(pvcName)},
			}, nil)
			mockPersistentVolumeClaims.On("Delete", pvcName, mock.Anything).Return(nil)
			secretName := "s3cr3t"
			mockSecrets.On("List", expectedListOptions).Return(&v1.SecretList{
				Items: []v1.Secret{secret(secretName)},
			}, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(nil)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(BeNil())
		})

		It("should successfully delete the service account, the PVC, the secret and the namespace itself", func() {
			options.RemoveNamespace = true
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(serviceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", serviceAccountName, mock.Anything).Return(nil)
			pvcName := "pvc-name"
			mockPersistentVolumeClaims.On("List", expectedListOptions).Return(&v1.PersistentVolumeClaimList{
				Items: []v1.PersistentVolumeClaim{persistentVolumeClaim(pvcName)},
			}, nil)
			mockPersistentVolumeClaims.On("Delete", pvcName, mock.Anything).Return(nil)
			secretName := "s3cr3t"
			mockSecrets.On("List", expectedListOptions).Return(&v1.SecretList{
				Items: []v1.Secret{secret(secretName)},
			}, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(nil)
			mockNamespaces.On("Delete", namespace, mock.Anything).Return(nil)

			err := kubectlClient.NamespaceCleanup(options)

			Expect(err).To(BeNil())
		})
	})
})

func makeKubectlClient(kubeClient kubernetes.Interface,
	kubeCtl kubectl.KubeCtl,
	kustomizer kustomize.Kustomizer) KubectlClient {
	return &kubectlClient{
		kubeClient: kubeClient,
		kubeCtl:    kubeCtl,
		kustomizer: kustomizer,
	}
}

func notFound() *errors.StatusError {
	return errors.NewNotFound(schema.GroupResource{}, "")
}

func namedAndLabelled(name string, labels map[string]string) func(sa *v1.ServiceAccount) bool {
	return func(sa *v1.ServiceAccount) bool {
		return sa.Name == name && reflect.DeepEqual(sa.Labels, labels)
	}
}

func named(name string) func(sa *v1.ServiceAccount) bool {
	return func(sa *v1.ServiceAccount) bool {
		return sa.Name == name
	}
}

func key(key string) func(dict map[string]string) bool {
	return func(dict map[string]string) bool {
		for k := range dict {
			if key == k {
				return true
			}
		}
		return false
	}
}

func urlPath(path string) func(url *url.URL) bool {
	return func(url *url.URL) bool {
		return url.Path == path
	}
}

func contentsOf(path string) []byte {
	absolutePath := unsafeAbs(path)
	bytes, err := ioutil.ReadFile(absolutePath)
	if err != nil {
		panic(err)
	}
	return bytes
}

func unsafeAbs(path string) string {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return absolutePath
}

func serviceAccount(name string) v1.ServiceAccount {
	return v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func persistentVolumeClaim(name string) v1.PersistentVolumeClaim {
	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func secret(name string) v1.Secret {
	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}
