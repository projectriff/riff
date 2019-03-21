/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http/httptest"
	"net/http"
)

var _ = Describe("HealthHandler", func() {
	var (
		mockResponseWriter *httptest.ResponseRecorder
		req                *http.Request
	)

	BeforeEach(func() {
		req = httptest.NewRequest("GET", "http://example.com", nil)
		req.URL.Path = "/health"
		mockResponseWriter = httptest.NewRecorder()

	})

	JustBeforeEach(func() {
		healthHandler(mockResponseWriter, req)
	})

	It("should return Ok", func() {
		resp := mockResponseWriter.Result()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
})
