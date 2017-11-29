/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http

import (
	"bytes"
	retry "github.com/giantswarm/retry-go"
	dispatcher "github.com/projectriff/function-sidecar/pkg/dispatcher"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

const UseTimeout = 10000000 // "Infinite" number of retries to override default and use the Timeout approach instead
const ConnectionAttemptTimeout = 1 * time.Minute
const ConnectionAttemptInterval = 100 * time.Millisecond

type httpDispatcher struct {
}

func (httpDispatcher) Dispatch(in *dispatcher.Message) (*dispatcher.Message, error) {
	slice := in.Payload.([]byte)

	client := http.Client{
		Timeout: time.Duration(60 * time.Second),
	}
	contentType := in.Headers.GetOrDefault("Content-Type", "application/octet-stream").(string)
	req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewReader(slice))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	if accept, ok := in.Headers["Accept"]; ok {
		req.Header.Add("Accept", accept.(string))
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Error invoking http://localhost:8080: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response %v\n", err)
		return nil, err
	}

	return &dispatcher.Message{Payload: out, Headers: flatten(resp.Header)}, nil
}

// http headers are a multi value map, dispatcher.Headers is single values (but with interface{} value)
// this function flattens headers which are mono-valued
func flatten(httpHeaders http.Header) dispatcher.Headers {
	result := make(map[string]interface{}, len(httpHeaders))
	for k, v := range httpHeaders {
		if len(v) == 1 {
			result[k] = v[0]
		} else if len(v) > 1 {
			result[k] = v
		}
	}
	return result
}

func NewHttpDispatcher() dispatcher.SynchDispatcher {
	attemptDial := func() error {
		log.Println("Waiting for function to accept connection on localhost:8080")
		_, err := net.Dial("tcp", "localhost:8080")
		return err
	}

	err := retry.Do(attemptDial,
		retry.Timeout(ConnectionAttemptTimeout),
		retry.Sleep(ConnectionAttemptInterval),
		retry.MaxTries(UseTimeout))
	if err != nil {
		panic(err)
	}
	return httpDispatcher{}
}
