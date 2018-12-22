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
	"k8s.io/kubernetes/pkg/kubectl"
	. "github.com/onsi/ginkgo"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks/mockbuilder"
	"github.com/projectriff/riff/pkg/core/vendor_mocks"
	"github.com/projectriff/riff/pkg/core/vendor_mocks/mockserving"
	"github.com/projectriff/riff/pkg/test_support"
	"github.com/stretchr/testify/mock"
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
		testService          *v1alpha1.Service
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

	Describe("CRD prototype", func() {
		Context("the real thing", func() {
			kubeCtl := kubectl.
			core.NewKubectlClient(kubeCtl)

		})
	})
})