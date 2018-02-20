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
	"github.com/projectriff/message-transport/pkg/message"
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

func (httpDispatcher) Dispatch(in message.Message) (message.Message, error) {
	client := http.Client{
		Timeout: time.Duration(60 * time.Second),
	}
	req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewReader(in.Payload()))
	if err != nil {
		log.Printf("Error creating POST request to http://localhost:8080: %v", err)
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

	result := message.NewMessage(out, make(map[string][]string))
	propagateOutgoingHeaders(resp, result)
	return result, nil
}

func propagateIncomingHeaders(message message.Message, request *http.Request) {
	for h, ss := range message.Headers() {
		for _, s := range ss {
			request.Header.Add(h, s)
		}
	}
}

func propagateOutgoingHeaders(resp *http.Response, message message.Message) {
	for h, v := range resp.Header {
		message.Headers()[h] = v
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
