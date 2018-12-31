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
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks/mockbuilder"
	"github.com/projectriff/riff/pkg/core/vendor_mocks"
	"github.com/projectriff/riff/pkg/core/vendor_mocks/mockserving"
	"github.com/projectriff/riff/pkg/test_support"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"path/filepath"
)

var _ = Describe("Function", func() {

	var (
		client               core.Client
		mockClientConfig     *vendor_mocks.ClientConfig
		mockBuilder          *mockbuilder.Builder
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
		mockBuilder = &mockbuilder.Builder{}
		mockServing = &mockserving.Interface{}
		mockServingV1alpha1 = &mockserving.ServingV1alpha1Interface{}
		mockServiceInterface = &mockserving.ServiceInterface{}
		mockServing.On("ServingV1alpha1").Return(mockServingV1alpha1)
		mockServingV1alpha1.On("Services", mock.Anything).Return(mockServiceInterface)
		testService = &v1alpha1.Service{}
		workDir = test_support.CreateTempDir()
		mockClientConfig.On("Namespace").Return("default", false, nil)
		client = core.NewClient(mockClientConfig, nil, nil, mockServing)
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
			service, err = client.CreateFunction(mockBuilder, createFunctionOptions, ioutil.Discard)
		})

		Context("when building locally", func() {
			BeforeEach(func() {
				mockServiceInterface.On("Create", mock.Anything).Run(func(args mock.Arguments) {
					createdService = args.Get(0).(*v1alpha1.Service)
				}).Return(testService, nil)
				mockServiceInterface.On("Get", mock.Anything, mock.Anything).
					Return(nil, notFound())
			})

			BeforeEach(func() {
				createFunctionOptions.LocalPath = workDir
				mockBuilder.On("Build", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			})

			Context("when buildpack and run images are provided", func() {
				BeforeEach(func() {
					createFunctionOptions.BuildpackImage = "some/buildpack"
					createFunctionOptions.RunImage = "some/run"
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
					// The returned service should be the input to service create, not the output.
					Expect(service).To(Equal(createdService))
				})
			})

			Context("when buildpack image is omitted", func() {
				BeforeEach(func() {
					createFunctionOptions.BuildpackImage = ""
					createFunctionOptions.RunImage = "some/run"
				})

				It("should return a suitable error", func() {
					Expect(err).To(MatchError("unable to build function locally: buildpack image not specified"))
				})
			})

			Context("when run image is omitted", func() {
				BeforeEach(func() {
					createFunctionOptions.BuildpackImage = "some/buildpack"
					createFunctionOptions.RunImage = ""
				})

				It("should return a suitable error", func() {
					Expect(err).To(MatchError("unable to build function locally: run image not specified"))
				})
			})

			Context("when riff.toml is already present", func() {
				BeforeEach(func() {
					if err := ioutil.WriteFile(filepath.Join(workDir, "riff.toml"), []byte{}, os.FileMode(0400));
						err != nil {
						panic(err)
					}
				})

				It("should fail", func() {
					msg := "found riff.toml file in local path. Please delete this file and let the CLI create it from flags"
					Expect(err).To(MatchError(msg))
				})
			})

			Context("when riff.toml is not initially present", func() {
				BeforeEach(func() {
					createFunctionOptions.BuildpackImage = "some/buildpack"
					createFunctionOptions.RunImage = "some/run"
				})

				It("should clean up created riff.toml upon function creation", func() {
					Expect(err).To(Not(HaveOccurred()))
					Expect(test_support.FileExists(filepath.Join(workDir, "riff.toml"))).To(BeFalse(),
						"expected riff.toml to be deleted upon function creation completion")
				})
			})
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
			err = client.UpdateFunction(mockBuilder, updateFunctionOptions, ioutil.Discard)
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

			Context("when buildpack and run images are provided", func() {
				BeforeEach(func() {
					testService.Annotations = map[string]string{"riff.projectriff.io-buildpack-buildImage": "some/buildpack",
						"riff.projectriff.io-buildpack-runImage": "some/run"}
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when buildpack image is omitted", func() {
				BeforeEach(func() {
					testService.Annotations = map[string]string{"riff.projectriff.io-buildpack-runImage": "some/run"}
				})

				It("should return a suitable error", func() {
					Expect(err).To(MatchError("unable to build function locally: buildpack image not specified"))
				})
			})

			Context("when run image is omitted", func() {
				BeforeEach(func() {
					testService.Annotations = map[string]string{"riff.projectriff.io-buildpack-buildImage": "some/buildpack"}
				})

				It("should return a suitable error", func() {
					Expect(err).To(MatchError("unable to build function locally: run image not specified"))
				})
			})
		})

		Context("when building in cluster", func() {
			BeforeEach(func() {
				functionName := "somefun"
				mockServiceInterface.On("Get", mock.Anything, mock.Anything).Return(testService, nil)
				testService.Spec = v1alpha1.ServiceSpec{
					RunLatest: &v1alpha1.RunLatestType{
						Configuration: v1alpha1.ConfigurationSpec{
							Build: &v1alpha1.RawExtension{}, // non-nil build === cluster-built
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
})

func notFound() *errors.StatusError {
	return errors.NewNotFound(schema.GroupResource{}, "")
}
