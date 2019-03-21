/*
 * Copyright 2019 The original author or authors
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

package test_support

import (
	"net"
	"net/http"
	"time"
)

type HttpResponse struct {
	StatusCode int
	Content    []byte
	Headers    map[string]string
}

func Serve(listener net.Listener, response HttpResponse) error {
	return ServeSlow(listener, response, 0)
}

func ServeSlow(listener net.Listener, response HttpResponse, delay time.Duration) error {
	err := http.Serve(listener, http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
		time.Sleep(delay)
		responseWriter.WriteHeader(statusCodeOrDefault(response, http.StatusOK))
		for headerKey, headerValue := range response.Headers {
			responseWriter.Header().Add(headerKey, headerValue)
		}
		_, _ = responseWriter.Write(response.Content)
	}))
	if err != nil {
		return err
	}
	return nil
}

func statusCodeOrDefault(response HttpResponse, defaultCode int) int {
	code := response.StatusCode
	if code == 0 {
		return defaultCode
	}
	return code
}
