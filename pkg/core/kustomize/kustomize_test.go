/*
 * Copyright 2019 The original author or authors
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

package kustomize_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core/kustomize"
	"github.com/projectriff/riff/pkg/test_support"
	"net"
	"net/url"
	"time"
)

var _ = Describe("Kustomize wrapper", func() {

	var (
		initLabels              map[string]string
		initialResourceContent  string
		httpResponse            test_support.HttpResponse
		expectedResourceContent string
		kustomizer              kustomize.Kustomizer
		timeout                 time.Duration
		workDir                 string
	)

	BeforeEach(func() {
		initLabels = map[string]string{"created-by": "riff", "because": "we-can"}
		initialResourceContent = `kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: riff-cnb-cache
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi`
		httpResponse = test_support.HttpResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/octet-stream"},
			Content:    []byte(initialResourceContent),
		}
		expectedResourceContent = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  labels:
    because: we-can
    created-by: riff
  name: riff-cnb-cache
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 8Gi
`
		timeout = 500 * time.Millisecond
		kustomizer = kustomize.MakeKustomizer(timeout)
		workDir = test_support.CreateTempDir()
	})

	AfterEach(func() {
		test_support.CleanupDirs(GinkgoT(), workDir)
	})

	It("customizes remote resources with provided labels", func() {
		resourceListener, _ := net.Listen("tcp", "127.0.0.1:0")
		go func() {
			err := test_support.Serve(resourceListener, httpResponse)
			Expect(err).NotTo(HaveOccurred())
		}()
		resourceUrl := unsafeParseUrl(fmt.Sprintf("http://%s/%s", resourceListener.Addr().String(), "pvc.yaml"))

		result, err := kustomizer.ApplyLabels(resourceUrl, initLabels)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(Equal(expectedResourceContent))
	})

	It("customizes local resources with provided labels", func() {
		file := test_support.CreateFile(workDir, "pvc.yaml", initialResourceContent)
		resourceUrl := unsafeParseUrl(test_support.FileURL(test_support.AbsolutePath(file)))

		result, err := kustomizer.ApplyLabels(resourceUrl, initLabels)

		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(Equal(expectedResourceContent))
	})

	It("fails on unsupported scheme", func() {
		resourceUrl := unsafeParseUrl("ftp://127.0.0.1/goodluck.yaml")

		_, err := kustomizer.ApplyLabels(resourceUrl, initLabels)

		Expect(err).To(MatchError("unsupported scheme in ftp://127.0.0.1/goodluck.yaml: ftp"))
	})

	It("fails if the resource is not reachable", func() {
		_, err := kustomizer.ApplyLabels(unsafeParseUrl("http://localhost:12345/nope.yaml"), initLabels)

		Expect(err).To(SatisfyAll(
			Not(BeNil()),
			BeAssignableToTypeOf(&url.Error{})))
	})

	It("fails if fetching the resource takes too long", func() {
		resourceListener, _ := net.Listen("tcp", "127.0.0.1:0")
		go func() {
			err := test_support.ServeSlow(resourceListener, httpResponse, 3*timeout)
			Expect(err).NotTo(HaveOccurred())
		}()
		resourceUrl := unsafeParseUrl(fmt.Sprintf("http://%s/%s", resourceListener.Addr().String(), "pvc.yaml"))

		_, err := kustomizer.ApplyLabels(resourceUrl, initLabels)

		Expect(err).To(SatisfyAll(
			Not(BeNil()),
			BeAssignableToTypeOf(&url.Error{})))
	})
})

func unsafeParseUrl(raw string) *url.URL {
	result, err := url.Parse(raw)
	if err != nil {
		Expect(err).NotTo(HaveOccurred())
	}
	return result
}
