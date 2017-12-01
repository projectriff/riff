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
	"github.com/giantswarm/retry-go"
	"github.com/projectriff/function-sidecar/pkg/dispatcher"
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
	req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewReader(slice))
	if err != nil {
		return nil, err
	}
	propagateIncomingHeaders(in, req)

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

	result := dispatcher.Message{Payload: out, Headers: make(map[string]interface{})}
	propagateOutgoingHeaders(resp, &result)
	return &result, nil
}

func propagateIncomingHeaders(message *dispatcher.Message, request *http.Request) {
	for h, v := range message.Headers {
		switch value := v.(type) {
		case string:
			request.Header.Add(h, value)
		case []string:
			for _, s := range value {
				request.Header.Add(h, s)
			}
		}
	}
}

func propagateOutgoingHeaders(resp *http.Response, message *dispatcher.Message) {
	for h, v := range resp.Header {
		// http headers are a multi value map, dispatcher.Headers is single values (but with interface{} value)
		// this function flattens headers which are mono-valued
		if len(v) == 1 {
			message.Headers[h] = v[0]
		} else if len(v) > 1 {
			message.Headers[h] = v
		}
	}
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
