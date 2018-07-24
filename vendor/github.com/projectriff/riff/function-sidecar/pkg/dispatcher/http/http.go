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
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/giantswarm/retry-go"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/riff/message-transport/pkg/message"
)

const UseTimeout = 10000000 // "Infinite" number of retries to override default and use the Timeout approach instead
const ConnectionAttemptTimeout = 1 * time.Minute
const ConnectionAttemptInterval = 100 * time.Millisecond

var headerBlacklist = map[string]*void{
	"correlationid": &void{},
}

type void struct{}

type httpDispatcher struct {
	port int
}

func (hd httpDispatcher) Dispatch(in message.Message) (message.Message, error) {
	client := http.Client{
		Timeout: time.Duration(60 * time.Second),
	}
	url := fmt.Sprintf("http://localhost:%d", hd.port)
	req, err := http.NewRequest("POST", url, bytes.NewReader(in.Payload()))
	if err != nil {
		log.Printf("Error creating POST request to %s: %v", url, err)
		return nil, err
	}
	propagateIncomingHeaders(in, req)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error invoking %s: %v", url, err)
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
		if headerBlacklist[strings.ToLower(h)] == nil {
			for _, s := range ss {
				request.Header.Add(h, s)
			}
		}
	}
}

func propagateOutgoingHeaders(resp *http.Response, message message.Message) {
	for h, v := range resp.Header {
		if headerBlacklist[strings.ToLower(h)] == nil {
			message.Headers()[h] = v
		}
	}
}

func NewHttpDispatcher(port int) dispatcher.SynchDispatcher {
	attemptDial := func() error {
		log.Printf("Waiting for function to accept connection on localhost:%d\n", port)
		_, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
		return err
	}

	err := retry.Do(attemptDial,
		retry.Timeout(ConnectionAttemptTimeout),
		retry.Sleep(ConnectionAttemptInterval),
		retry.MaxTries(UseTimeout))
	if err != nil {
		panic(err)
	}
	return httpDispatcher{
		port,
	}
}
