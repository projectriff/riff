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

package core_test

import (
	"github.com/buildpack/pack"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks/mockbuildfactory"
	"github.com/projectriff/riff/pkg/core/vendor_mocks"
	"github.com/projectriff/riff/pkg/core/vendor_mocks/mockserving"
	"github.com/projectriff/riff/pkg/test_support"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Function", func() {

	var (
		client               core.Client
		mockClientConfig     *vendor_mocks.ClientConfig
		mockBuildFactory     *mockbuildfactory.BuildFactory
		mockBuild            *mockbuildfactory.Build
		mockServing          *mockserving.Interface
		mockServingV1alpha1  *mockserving.ServingV1alpha1Interface
		mockServiceInterface *mockserving.ServiceInterface
		workDir              string
		service              *v1alpha1.Service
		testService          *v1alpha1.Service
		err                  error
	)

	BeforeEach(func() {
		mockClientConfig = &vendor_mocks.ClientConfig{}
		mockBuildFactory = &mockbuildfactory.BuildFactory{}
		mockBuild = &mockbuildfactory.Build{}
		mockServing = &mockserving.Interface{}
		mockServingV1alpha1 = &mockserving.ServingV1alpha1Interface{}
		mockServiceInterface = &mockserving.ServiceInterface{}
		mockServing.On("ServingV1alpha1").Return(mockServingV1alpha1)
		mockServingV1alpha1.On("Services", mock.Anything).Return(mockServiceInterface)
		testService = &v1alpha1.Service{}
		workDir = test_support.CreateTempDir()
		mockClientConfig.On("Namespace").Return("default", false, nil)
		client = core.NewClient(mockClientConfig, nil, nil, mockServing, mockBuildFactory)
	})

	AfterEach(func() {
		test_support.CleanupDirs(GinkgoT(), workDir)
	})

	Describe("CreateFunction", func() {
		var (
			createFunctionOptions core.CreateFunctionOptions
			createdService        *v1alpha1.Service
		)

		BeforeEach(func() {
			mockServiceInterface.On("Create", mock.Anything).Run(func(args mock.Arguments) {
				createdService = args.Get(0).(*v1alpha1.Service)
			}).Return(testService, nil)
		})

		JustBeforeEach(func() {
			service, err = client.CreateFunction(createFunctionOptions, ioutil.Discard)
		})

		Context("when building locally", func() {
			var buildConfig *pack.BuildFlags

			BeforeEach(func() {
				createFunctionOptions.LocalPath = workDir
				mockBuildFactory.On("BuildConfigFromFlags", mock.Anything).Return(mockBuild, nil).Run(func(args mock.Arguments) {
					buildConfig = args.Get(0).(*pack.BuildFlags)
				})
				mockBuild.On("Run").Return(nil)
			})

			Context("when buildpack and run images are not both located on dev.local", func() {
				BeforeEach(func() {
					createFunctionOptions.BuildpackImage = "dev.local/buildpack"
					createFunctionOptions.RunImage = "some/run"
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
					// The returned service should be the input to service create, not the output.
					Expect(service).To(Equal(createdService))
				})

				It("should build not specifying 'no pull'", func() {
					Expect(buildConfig.NoPull).To(BeFalse())
				})
			})

			Context("when buildpack and run images are located on dev.local", func() {
				BeforeEach(func() {
					createFunctionOptions.BuildpackImage = "dev.local/buildpack"
					createFunctionOptions.RunImage = "dev.local/run"
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
					// The returned service should be the input to service create, not the output.
					Expect(service).To(Equal(createdService))
				})

				It("should build specifying 'no pull'", func() {
					Expect(buildConfig.NoPull).To(BeTrue())
				})
			})
		})
	})

	Describe("UpdateFunction", func() {
		var (
			updateFunctionOptions core.UpdateFunctionOptions
		)

		BeforeEach(func() {
			mockServiceInterface.On("Get", mock.Anything, mock.Anything).Return(testService, nil)
			testService.Spec = v1alpha1.ServiceSpec{
				RunLatest: &v1alpha1.RunLatestType{
					Configuration: v1alpha1.ConfigurationSpec{
						Build: nil,
						RevisionTemplate: v1alpha1.RevisionTemplateSpec{
							ObjectMeta: v1.ObjectMeta{
								Labels: map[string]string{"riff.projectriff.io/function": "somefun"},
							},
						},
					},
				},
			}
			mockServiceInterface.On("Update", mock.Anything).Return(testService, nil)
		})

		JustBeforeEach(func() {
			err = client.UpdateFunction(updateFunctionOptions, ioutil.Discard)
		})

		Context("when building locally", func() {
			var buildConfig *pack.BuildFlags

			BeforeEach(func() {
				updateFunctionOptions.LocalPath = workDir
				mockBuildFactory.On("BuildConfigFromFlags", mock.Anything).Return(mockBuild, nil).Run(func(args mock.Arguments) {
					buildConfig = args.Get(0).(*pack.BuildFlags)
				})
				mockBuild.On("Run").Return(nil)
			})

			Context("when buildpack and run images are not both located on dev.local", func() {
				BeforeEach(func() {
					testService.Annotations = map[string]string{"riff.projectriff.io-buildpack-buildImage": "dev.local/buildpack",
						"riff.projectriff.io-buildpack-runImage": "some/run"}
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should build not specifying 'no pull'", func() {
					Expect(buildConfig.NoPull).To(BeFalse())
				})
			})

			Context("when buildpack and run images are located on dev.local", func() {
				BeforeEach(func() {
					testService.Annotations = map[string]string{"riff.projectriff.io-buildpack-buildImage": "dev.local/buildpack",
						"riff.projectriff.io-buildpack-runImage": "dev.local/run"}
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should build specifying 'no pull'", func() {
					Expect(buildConfig.NoPull).To(BeTrue())
				})
			})
		})
	})
})
