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
	"context"
	"io/ioutil"

	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks/mockbuilder"
	"github.com/projectriff/riff/pkg/core/vendor_mocks"
	"github.com/projectriff/riff/pkg/core/vendor_mocks/mockbuild"
	"github.com/projectriff/riff/pkg/core/vendor_mocks/mockserving"
	"github.com/projectriff/riff/pkg/test_support"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Function", func() {

	var (
		client                            core.Client
		mockClientConfig                  *vendor_mocks.ClientConfig
		mockBuilder                       *mockbuilder.Builder
		mockServing                       *mockserving.Interface
		mockServingV1alpha1               *mockserving.ServingV1alpha1Interface
		mockServiceInterface              *mockserving.ServiceInterface
		mockRevisionInterface             *mockserving.RevisionInterface
		mockBuild                         *mockbuild.Interface
		mockBuildV1alpha1                 *mockbuild.BuildV1alpha1Interface
		mockClusterBuildTemplateInterface *mockbuild.ClusterBuildTemplateInterface
		workDir                           string
		service                           *v1alpha1.Service
		testService                       *v1alpha1.Service
		revision                          *v1alpha1.Revision
		testRevision                      *v1alpha1.Revision
		cache                             *corev1.PersistentVolumeClaim
		err                               error
	)

	BeforeEach(func() {
		mockClientConfig = &vendor_mocks.ClientConfig{}
		mockBuilder = &mockbuilder.Builder{}
		mockServing = &mockserving.Interface{}
		mockServingV1alpha1 = &mockserving.ServingV1alpha1Interface{}
		mockServiceInterface = &mockserving.ServiceInterface{}
		mockRevisionInterface = &mockserving.RevisionInterface{}
		mockServing.On("ServingV1alpha1").Return(mockServingV1alpha1)
		mockServingV1alpha1.On("Services", mock.Anything).Return(mockServiceInterface)
		mockServingV1alpha1.On("Revisions", mock.Anything).Return(mockRevisionInterface)
		mockBuild = &mockbuild.Interface{}
		mockBuildV1alpha1 = &mockbuild.BuildV1alpha1Interface{}
		mockClusterBuildTemplateInterface = &mockbuild.ClusterBuildTemplateInterface{}
		mockBuild.On("BuildV1alpha1").Return(mockBuildV1alpha1)
		mockBuildV1alpha1.On("ClusterBuildTemplates").Return(mockClusterBuildTemplateInterface)
		testService = &v1alpha1.Service{}
		workDir = test_support.CreateTempDir()
		mockClientConfig.On("Namespace").Return("default", false, nil)
		client = core.NewClient(mockClientConfig, nil, mockServing, mockBuild, nil, nil)

		builderImage := "projectriff/builder:testing"
		mockClusterBuildTemplateInterface.On("Get", mock.Anything, mock.Anything).
			Return(&build.ClusterBuildTemplate{
				Spec: build.BuildTemplateSpec{
					Parameters: []build.ParameterSpec{
						{Name: "BUILDER_IMAGE", Default: &builderImage},
					},
				},
			}, nil)
	})

	AfterEach(func() {
		test_support.CleanupDirs(GinkgoT(), workDir)
	})

	Describe("CreateFunction", func() {
		var (
			createFunctionOptions core.CreateFunctionOptions
			createdService        *v1alpha1.Service
		)

		JustBeforeEach(func() {
			service, revision, cache, err = client.CreateFunction(context.TODO(), mockBuilder, createFunctionOptions, ioutil.Discard)
		})

		Context("when building locally", func() {
			BeforeEach(func() {
				mockServiceInterface.On("Get", mock.Anything, mock.Anything).
					Return(nil, notFound()).Once()
				mockServiceInterface.On("Create", mock.Anything).Run(func(args mock.Arguments) {
					createdService = args.Get(0).(*v1alpha1.Service)
				}).Return(testService, nil).Once()
				mockServiceInterface.On("Get", mock.Anything, mock.Anything).
					Return(testService, nil).Once()

				mockRevisionInterface.On("Get", testService.Status.LatestCreatedRevisionName, mock.Anything).
					Return(testRevision, nil).Once()
			})

			BeforeEach(func() {
				createFunctionOptions.LocalPath = workDir
				mockBuilder.On("Build", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})

			Context("when buildpack image is provided", func() {
				BeforeEach(func() {
					createFunctionOptions.BuildpackImage = "some/buildpack"
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
					// The returned service should be the input to service create, not the output.
					Expect(service).To(Equal(createdService))
					Expect(revision).To(BeNil())
					Expect(cache).To(BeNil())
				})
			})

			Context("when buildpack image is omitted", func() {
				BeforeEach(func() {
					createFunctionOptions.BuildpackImage = ""
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
					// The returned service should be the input to service create, not the output.
					Expect(service).To(Equal(createdService))
					Expect(revision).To(BeNil())
					Expect(cache).To(BeNil())
				})
			})

			// TODO add tests for --wait and --verbose
		})

		Context("when a service with the same name already exists", func() {
			BeforeEach(func() {
				functionName := "notAvailable"
				createFunctionOptions.Name = functionName
				existingService := &v1alpha1.Service{}
				existingService.Name = functionName
				mockServiceInterface.On("Get", "notAvailable", mock.Anything).
					Return(existingService, nil)
			})

			It("should fail early", func() {
				Expect(err).To(MatchError(`service 'notAvailable' already exists in namespace 'default'`))
			})
		})
	})

	Describe("UpdateFunction", func() {
		var (
			updateFunctionOptions core.UpdateFunctionOptions
		)

		BeforeEach(func() {
			mockServiceInterface.On("Update", mock.Anything).Return(testService, nil)
		})

		JustBeforeEach(func() {
			err = client.UpdateFunction(context.TODO(), mockBuilder, updateFunctionOptions, ioutil.Discard)
		})

		Context("when building locally", func() {
			BeforeEach(func() {
				mockServiceInterface.On("Get", mock.Anything, mock.Anything).Return(testService, nil)
				testService.Spec = v1alpha1.ServiceSpec{
					RunLatest: &v1alpha1.RunLatestType{
						Configuration: v1alpha1.ConfigurationSpec{
							Build: nil, // nil build === local build
							RevisionTemplate: v1alpha1.RevisionTemplateSpec{
								ObjectMeta: v1.ObjectMeta{
									Labels: map[string]string{"riff.projectriff.io/function": "somefun"},
								},
							},
						},
					},
				}
				updateFunctionOptions.LocalPath = workDir
				mockBuilder.On("Build", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when building in cluster", func() {
			BeforeEach(func() {
				functionName := "somefun"
				mockServiceInterface.On("Get", mock.Anything, mock.Anything).Return(testService, nil)
				testService.Spec = v1alpha1.ServiceSpec{
					RunLatest: &v1alpha1.RunLatestType{
						Configuration: v1alpha1.ConfigurationSpec{
							Build: &v1alpha1.RawExtension{ // non-nil build === cluster-built
								Object: &build.Build{
									TypeMeta: v1.TypeMeta{
										APIVersion: "build.knative.dev/v1alpha1",
										Kind:       "Build",
									},
									Spec: build.BuildSpec{},
								},
							},
							RevisionTemplate: v1alpha1.RevisionTemplateSpec{
								ObjectMeta: v1.ObjectMeta{
									Labels: map[string]string{"riff.projectriff.io/function": functionName},
								},
							},
						},
					},
				}
				updateFunctionOptions.LocalPath = "should-trigger-failure"
				updateFunctionOptions.Name = functionName
			})

			It("should fail if a local path is provided", func() {
				Expect(err).To(MatchError(`unable to proceed: local path specified for cluster-built service named "somefun"`))
			})
		})
	})

	Describe("BuildFunction", func() {
		var (
			buildFunctionOptions core.BuildFunctionOptions
		)

		JustBeforeEach(func() {
			err = client.BuildFunction(context.TODO(), mockBuilder, buildFunctionOptions, ioutil.Discard)
		})

		Context("when building locally", func() {
			BeforeEach(func() {
				buildFunctionOptions.LocalPath = workDir
				mockBuilder.On("Build", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})
})

func notFound() *errors.StatusError {
	return errors.NewNotFound(schema.GroupResource{}, "")
}
