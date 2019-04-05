/*
 * Copyright 2018-2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core_test

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/projectriff/riff/pkg/core"
	mockkustomize "github.com/projectriff/riff/pkg/core/kustomize/mocks"
	"github.com/projectriff/riff/pkg/core/vendor_mocks"
	"github.com/projectriff/riff/pkg/env"
	mockkubectl "github.com/projectriff/riff/pkg/kubectl/mocks"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("namespace", func() {

	var (
		client              core.Client
		kubeClient          *vendor_mocks.Interface
		kubeCtl             *mockkubectl.KubeCtl
		mockCore            *vendor_mocks.CoreV1Interface
		mockNamespaces      *vendor_mocks.NamespaceInterface
		mockServiceAccounts *vendor_mocks.ServiceAccountInterface
		mockConfigMaps      *vendor_mocks.ConfigMapInterface
		mockSecrets         *vendor_mocks.SecretInterface
		mockKustomizer      *mockkustomize.Kustomizer
		manifests           map[string]*core.Manifest
	)

	JustBeforeEach(func() {
		kubeClient = new(vendor_mocks.Interface)
		kubeCtl = new(mockkubectl.KubeCtl)
		mockCore = new(vendor_mocks.CoreV1Interface)
		mockNamespaces = new(vendor_mocks.NamespaceInterface)
		mockServiceAccounts = new(vendor_mocks.ServiceAccountInterface)
		mockConfigMaps = new(vendor_mocks.ConfigMapInterface)
		mockSecrets = new(vendor_mocks.SecretInterface)
		mockKustomizer = new(mockkustomize.Kustomizer)
		manifests = map[string]*core.Manifest{}

		kubeClient.On("CoreV1").Return(mockCore)
		mockCore.On("Namespaces").Return(mockNamespaces)
		mockCore.On("ServiceAccounts", mock.Anything).Return(mockServiceAccounts)
		mockCore.On("ConfigMaps", mock.Anything).Return(mockConfigMaps)
		mockCore.On("Secrets", mock.Anything).Return(mockSecrets)

		client = core.NewClient(nil, kubeClient, nil, nil, nil, kubeCtl, mockKustomizer)
	})

	AfterEach(func() {
		mockNamespaces.AssertExpectations(GinkgoT())
		mockServiceAccounts.AssertExpectations(GinkgoT())
		mockSecrets.AssertExpectations(GinkgoT())
		mockConfigMaps.AssertExpectations(GinkgoT())
		mockKustomizer.AssertExpectations(GinkgoT())
	})

	Describe("NamespaceInit", func() {

		It("should fail on wrong manifest", func() {
			options := core.NamespaceInitOptions{Manifest: "wrong"}
			err := client.NamespaceInit(manifests, options)
			Expect(err).To(MatchError(ContainSubstring("wrong: "))) // error message is quite different on Windows and macOS
		})

		It("should create namespace and sa if needed", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "fixtures/empty.yaml",
				NamespaceName: "foo",
				SecretName:    "push-credentials",
			}
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(nil, notFound())
			mockNamespaces.On("Create", namespace).Return(namespace, nil)
			mockSecrets.On("Get", "push-credentials", metav1.GetOptions{}).Return(&v1.Secret{}, nil)
			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			labels := map[string]string{
				"projectriff.io/installer": env.Cli.Name,
				"projectriff.io/version":   env.Cli.Version,
			}
			mockServiceAccounts.On("Create", mock.MatchedBy(namedAndLabelled(core.BuildServiceAccountName, labels))).Return(serviceAccount, nil)

			configMap := &v1.ConfigMap{}
			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig(""))).Return(configMap, nil)

			err := client.NamespaceInit(manifests, options)

			Expect(err).To(Not(HaveOccurred()))
		})

		It("should create secret for gcr", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "fixtures/empty.yaml",
				NamespaceName: "foo",
				GcrTokenPath:  "fixtures/gcr-creds",
				SecretName:    "push-credentials",
			}

			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(serviceAccount, nil)

			secret := &v1.Secret{}
			mockSecrets.On("Delete", "push-credentials", &metav1.DeleteOptions{}).Return(nil)
			mockSecrets.On("Create", mock.Anything).Run(func(args mock.Arguments) {
				s := args[0].(*v1.Secret)
				Expect(s.StringData).To(HaveKeyWithValue("username", "_json_key"))
				var expectedPassword string
				if runtime.GOOS == "windows" {
					expectedPassword = "{ \"project_id\": \"gcp-project-id\" }\r\n"
				} else {
					expectedPassword = "{ \"project_id\": \"gcp-project-id\" }\n"
				}
				Expect(s.StringData).To(HaveKeyWithValue("password", expectedPassword))
				Expect(s.Labels).To(HaveLen(2))
				Expect(s.Labels["projectriff.io/installer"]).To(Equal(env.Cli.Name))
				Expect(s.Labels["projectriff.io/version"]).To(Equal(env.Cli.Version))
			}).Return(secret, nil)

			mockServiceAccounts.On("Update", mock.Anything).Run(func(args mock.Arguments) {
				sa := args[0].(*v1.ServiceAccount)
				Expect(sa.Secrets).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("push-credentials"),
				})))
			}).Return(serviceAccount, nil)

			configMap := &v1.ConfigMap{}
			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig("gcr.io/gcp-project-id"))).Return(configMap, nil)

			err := client.NamespaceInit(manifests, options)
			Expect(err).To(Not(HaveOccurred()))
		})

		Context("when dealing with Dockerhub", func() {

			var oldStdIn *os.File

			BeforeEach(func() {
				oldStdIn = os.Stdin
				creds, _ := os.Open("fixtures/registry-password")
				os.Stdin = creds
			})

			AfterEach(func() {
				os.Stdin = oldStdIn
			})

			It("should create secret for dockerhub", func() {

				options := core.NamespaceInitOptions{
					Manifest:      "fixtures/empty.yaml",
					NamespaceName: "foo",
					DockerHubId:   "roger",
					SecretName:    "push-credentials",
				}

				namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
				mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

				serviceAccount := &v1.ServiceAccount{}
				mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(serviceAccount, nil)

				secret := &v1.Secret{}
				mockSecrets.On("Delete", "push-credentials", &metav1.DeleteOptions{}).Return(nil)
				mockSecrets.On("Create", mock.Anything).Run(func(args mock.Arguments) {
					s := args[0].(*v1.Secret)
					Expect(s.StringData).To(HaveKeyWithValue("username", "roger"))
					Expect(s.StringData).To(HaveKeyWithValue("password", "s3cr3t"))
					Expect(s.Labels).To(HaveLen(2))
					Expect(s.Labels["projectriff.io/installer"]).To(Equal(env.Cli.Name))
					Expect(s.Labels["projectriff.io/version"]).To(Equal(env.Cli.Version))
				}).Return(secret, nil)

				mockServiceAccounts.On("Update", mock.Anything).Run(func(args mock.Arguments) {
					sa := args[0].(*v1.ServiceAccount)
					Expect(sa.Secrets).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Name": Equal("push-credentials"),
					})))
				}).Return(serviceAccount, nil)

				configMap := &v1.ConfigMap{}
				mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
				mockConfigMaps.On("Create", mock.MatchedBy(buildConfig("docker.io/roger"))).Return(configMap, nil)

				err := client.NamespaceInit(manifests, options)
				Expect(err).To(Not(HaveOccurred()))
			})
		})

		Context("when dealing with a basic auth registry", func() {

			var oldStdIn *os.File

			BeforeEach(func() {
				oldStdIn = os.Stdin
				creds, _ := os.Open("fixtures/registry-password")
				os.Stdin = creds
			})

			AfterEach(func() {
				os.Stdin = oldStdIn
			})

			It("should create a secret for the registry", func() {

				options := core.NamespaceInitOptions{
					Manifest:      "fixtures/empty.yaml",
					NamespaceName: "foo",
					Registry:      "https://registry.example.com",
					RegistryUser:  "roger",
					SecretName:    "push-credentials",
				}

				namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
				mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

				serviceAccount := &v1.ServiceAccount{}
				mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(serviceAccount, nil)

				secret := &v1.Secret{}
				mockSecrets.On("Delete", "push-credentials", &metav1.DeleteOptions{}).Return(nil)
				mockSecrets.On("Create", mock.Anything).Run(func(args mock.Arguments) {
					s := args[0].(*v1.Secret)
					Expect(s.ObjectMeta.Annotations).To(HaveKeyWithValue("build.knative.dev/docker-0", "https://registry.example.com"))
					Expect(s.StringData).To(HaveKeyWithValue("username", "roger"))
					Expect(s.StringData).To(HaveKeyWithValue("password", "s3cr3t"))
					Expect(s.Labels).To(HaveLen(2))
					Expect(s.Labels["projectriff.io/installer"]).To(Equal(env.Cli.Name))
					Expect(s.Labels["projectriff.io/version"]).To(Equal(env.Cli.Version))
				}).Return(secret, nil)

				mockServiceAccounts.On("Update", mock.Anything).Run(func(args mock.Arguments) {
					sa := args[0].(*v1.ServiceAccount)
					Expect(sa.Secrets).To(ContainElement(MatchFields(IgnoreExtras, Fields{
						"Name": Equal("push-credentials"),
					})))
				}).Return(serviceAccount, nil)

				configMap := &v1.ConfigMap{}
				mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
				mockConfigMaps.On("Create", mock.MatchedBy(buildConfig(""))).Return(configMap, nil)

				err := client.NamespaceInit(manifests, options)
				Expect(err).To(Not(HaveOccurred()))
			})
		})

		It("should run unauthenticated and still create a service account", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "fixtures/empty.yaml",
				NamespaceName: "foo",
				NoSecret:      true,
			}

			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(core.BuildServiceAccountName))).Return(serviceAccount, nil)

			configMap := &v1.ConfigMap{}
			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig(""))).Return(configMap, nil)

			err := client.NamespaceInit(manifests, options)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should set the default image prefix if defined", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "fixtures/empty.yaml",
				NamespaceName: "foo",
				NoSecret:      true,
				ImagePrefix:   "registry.example.com",
			}

			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(core.BuildServiceAccountName))).Return(serviceAccount, nil)

			configMap := &v1.ConfigMap{}
			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig("registry.example.com"))).Return(configMap, nil)

			err := client.NamespaceInit(manifests, options)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should apply label to namespace resources", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "stable",
				NamespaceName: "foo",
				NoSecret:      true,
			}
			namespaceResource := unsafeAbs("fixtures" + string(os.PathSeparator) + "initial_pvc.yaml")
			manifests["stable"] = &core.Manifest{
				Namespace: []string{namespaceResource},
			}
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(core.BuildServiceAccountName))).Return(serviceAccount, nil)

			configMap := &v1.ConfigMap{}
			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig(""))).Return(configMap, nil)

			customizedResourceContents := contentsOf("fixtures" + string(os.PathSeparator) + "kustom_pvc.yaml")
			mockKustomizer.On("ApplyLabels",
				mock.MatchedBy(urlPath(namespaceResource)),
				mock.MatchedBy(keys("projectriff.io/installer", "projectriff.io/version"))).
				Return(customizedResourceContents, nil)
			kubeCtl.On("ExecStdin", []string{"apply", "-n", "foo", "-f", "-"}, &customizedResourceContents).
				Return("done!", nil)

			err := client.NamespaceInit(manifests, options)

			Expect(err).To(Not(HaveOccurred()))
		})

		It("should fail if the image prefix deletion fails", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "stable",
				NamespaceName: "foo",
				NoSecret:      true,
			}
			namespaceResource := unsafeAbs("fixtures/initial_pvc.yaml")
			manifests["stable"] = &core.Manifest{
				Namespace: []string{namespaceResource},
			}
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(core.BuildServiceAccountName))).Return(serviceAccount, nil)
			expectedError := fmt.Errorf("image prefix deletion failed")

			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig(""))).Return(nil, expectedError)

			err := client.NamespaceInit(manifests, options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should not fail if the image prefix deletion returns a not found error", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "stable",
				NamespaceName: "foo",
				NoSecret:      true,
			}
			namespaceResource := unsafeAbs("fixtures/initial_pvc.yaml")
			manifests["stable"] = &core.Manifest{
				Namespace: []string{namespaceResource},
			}
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(core.BuildServiceAccountName))).Return(serviceAccount, nil)

			configMap := &v1.ConfigMap{}
			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig(""))).Return(configMap, nil)

			customizedResourceContents := contentsOf("fixtures/kustom_pvc.yaml")
			mockKustomizer.On("ApplyLabels",
				mock.MatchedBy(urlPath(namespaceResource)),
				mock.MatchedBy(keys("projectriff.io/installer", "projectriff.io/version"))).
				Return(customizedResourceContents, nil)
			kubeCtl.On("ExecStdin", []string{"apply", "-n", "foo", "-f", "-"}, &customizedResourceContents).
				Return("done!", nil)

			err := client.NamespaceInit(manifests, options)

			Expect(err).To(BeNil())
		})

		It("should fail if the PVC label kustomization fails", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "stable",
				NamespaceName: "foo",
				NoSecret:      true,
			}
			namespaceResource := unsafeAbs("fixtures/initial_pvc.yaml")
			manifests["stable"] = &core.Manifest{
				Namespace: []string{namespaceResource},
			}
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(core.BuildServiceAccountName))).Return(serviceAccount, nil)

			configMap := &v1.ConfigMap{}
			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig(""))).Return(configMap, nil)

			expectedError := fmt.Errorf("kustomization failed")
			mockKustomizer.On("ApplyLabels",
				mock.MatchedBy(urlPath(namespaceResource)),
				mock.MatchedBy(keys("projectriff.io/installer", "projectriff.io/version"))).
				Return(nil, expectedError)

			err := client.NamespaceInit(manifests, options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should support local kubernetes configuration files", func() {
			options := core.NamespaceInitOptions{
				Manifest:      "fixtures/local-yaml/manifest.yaml",
				NamespaceName: "foo",
				NoSecret:      true,
			}

			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
			mockNamespaces.On("Get", "foo", mock.Anything).Return(namespace, nil)

			serviceAccount := &v1.ServiceAccount{}
			mockServiceAccounts.On("Get", core.BuildServiceAccountName, mock.Anything).Return(nil, notFound())
			mockServiceAccounts.On("Create", mock.MatchedBy(named(core.BuildServiceAccountName))).Return(serviceAccount, nil)

			configMap := &v1.ConfigMap{}
			mockConfigMaps.On("Get", core.BuildConfigMapName, mock.Anything).Return(nil, notFound())
			mockConfigMaps.On("Create", mock.MatchedBy(buildConfig(""))).Return(configMap, nil)

			resource := unsafeAbs("fixtures/local-yaml/buildtemplate.yaml")
			mockKustomizer.On("ApplyLabels",
				mock.MatchedBy(urlPath(resource)),
				mock.MatchedBy(keys("projectriff.io/installer", "projectriff.io/version"))).
				Return([]byte("customised content"), nil)

			kubeCtl.On("ExecStdin", []string{"apply", "-n", "foo", "-f", "-"}, mock.Anything).
				Return("done!", nil)

			err := client.NamespaceInit(manifests, options)
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	Describe("NamespaceCleanup", func() {

		var (
			namespace           string
			options             core.NamespaceCleanupOptions
			expectedListOptions metav1.ListOptions
		)

		BeforeEach(func() {
			namespace = "foo"
			options = core.NamespaceCleanupOptions{
				NamespaceName: namespace,
			}
			expectedListOptions = metav1.ListOptions{LabelSelector: "projectriff.io/installer,projectriff.io/version"}
		})

		It("should fail if the service account list fails", func() {
			expectedError := fmt.Errorf("SA list failed")
			mockServiceAccounts.On("List", expectedListOptions).Return(nil, expectedError)

			err := client.NamespaceCleanup(options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should fail if the service account deletion fails", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(core.BuildServiceAccountName)},
			}, nil)
			expectedError := fmt.Errorf("SA deletion failed")
			mockServiceAccounts.On("Delete", core.BuildServiceAccountName, mock.Anything).Return(expectedError)

			err := client.NamespaceCleanup(options)

			Expect(err).To(MatchError(fmt.Sprintf("Unable to delete service account %s: %s", core.BuildServiceAccountName, expectedError.Error())))
		})

		It("should fail if the secret list fails", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(core.BuildServiceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", core.BuildServiceAccountName, mock.Anything).Return(nil)
			expectedError := fmt.Errorf("secret deletion failed")
			mockSecrets.On("List", expectedListOptions).Return(nil, expectedError)

			err := client.NamespaceCleanup(options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should fail if the secret deletion fails", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(core.BuildServiceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", core.BuildServiceAccountName, mock.Anything).Return(nil)
			secretName := "s3cr3t"
			mockSecrets.On("List", expectedListOptions).Return(&v1.SecretList{
				Items: []v1.Secret{secret(secretName)},
			}, nil)
			expectedError := fmt.Errorf("secret deletion failed")
			mockSecrets.On("Delete", secretName, mock.Anything).Return(expectedError)

			err := client.NamespaceCleanup(options)

			Expect(err).To(MatchError(fmt.Sprintf("Unable to delete secret %s: %s", secretName, expectedError.Error())))
		})

		It("should fail if the namespace deletion fails", func() {
			options.RemoveNamespace = true
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(core.BuildServiceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", core.BuildServiceAccountName, mock.Anything).Return(nil)
			secretName := "s3cr3t"
			mockSecrets.On("List", expectedListOptions).Return(&v1.SecretList{
				Items: []v1.Secret{secret(secretName)},
			}, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(nil)
			configMapName := "riff-build"
			mockConfigMaps.On("List", expectedListOptions).Return(&v1.ConfigMapList{
				Items: []v1.ConfigMap{configMap(configMapName)},
			}, nil)
			mockConfigMaps.On("Delete", core.BuildConfigMapName, mock.Anything).Return(nil)
			expectedError := fmt.Errorf("namespace deletion failed")
			mockNamespaces.On("Delete", namespace, mock.Anything).Return(expectedError)

			err := client.NamespaceCleanup(options)

			Expect(err).To(MatchError(expectedError))
		})

		It("should successfully delete the service account and the secret", func() {
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(core.BuildServiceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", core.BuildServiceAccountName, mock.Anything).Return(nil)
			secretName := "s3cr3t"
			mockSecrets.On("List", expectedListOptions).Return(&v1.SecretList{
				Items: []v1.Secret{secret(secretName)},
			}, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(nil)
			configMapName := "riff-build"
			mockConfigMaps.On("List", expectedListOptions).Return(&v1.ConfigMapList{
				Items: []v1.ConfigMap{configMap(configMapName)},
			}, nil)
			mockConfigMaps.On("Delete", core.BuildConfigMapName, mock.Anything).Return(nil)

			err := client.NamespaceCleanup(options)

			Expect(err).To(BeNil())
		})

		It("should successfully delete the service account, the secret and the namespace itself", func() {
			options.RemoveNamespace = true
			mockServiceAccounts.On("List", expectedListOptions).Return(&v1.ServiceAccountList{
				Items: []v1.ServiceAccount{serviceAccount(core.BuildServiceAccountName)},
			}, nil)
			mockServiceAccounts.On("Delete", core.BuildServiceAccountName, mock.Anything).Return(nil)
			secretName := "s3cr3t"
			mockSecrets.On("List", expectedListOptions).Return(&v1.SecretList{
				Items: []v1.Secret{secret(secretName)},
			}, nil)
			mockSecrets.On("Delete", secretName, mock.Anything).Return(nil)
			configMapName := "riff-build"
			mockConfigMaps.On("List", expectedListOptions).Return(&v1.ConfigMapList{
				Items: []v1.ConfigMap{configMap(configMapName)},
			}, nil)
			mockConfigMaps.On("Delete", core.BuildConfigMapName, mock.Anything).Return(nil)
			mockNamespaces.On("Delete", namespace, mock.Anything).Return(nil)

			err := client.NamespaceCleanup(options)

			Expect(err).To(BeNil())
		})
	})
})

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

func buildConfig(prefix string) func(cm *v1.ConfigMap) bool {
	return func(cm *v1.ConfigMap) bool {
		return cm.Data[core.DefaultImagePrefixKey] == prefix
	}
}

func keys(keys ...string) func(dict map[string]string) bool {
	return func(dict map[string]string) bool {
		checks := make(map[string]bool, len(dict))
		for k := range dict {
			checks[k] = true
		}
		result := true
		i := 0
		for i < len(keys) && result {
			result = result && checks[keys[i]]
			i++
		}
		return result
	}
}

func urlPath(path string) func(url *url.URL) bool {
	return func(url *url.URL) bool {
		if runtime.GOOS == "windows" && len(url.Scheme) == 1 {
			var drive string
			if url.Scheme == path[0:1] {
				drive = url.Scheme
			} else {
				drive = strings.ToUpper(url.Scheme)
			}
			return drive + url.String()[1:] == path
		} else {
			return url.Path == path
		}
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

func secret(name string) v1.Secret {
	return v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func configMap(name string) v1.ConfigMap {
	return v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}
