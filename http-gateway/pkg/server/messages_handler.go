/*
 * Copyright 2018 the original author or authors.
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

package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/projectriff/riff/message-transport/pkg/message"
)

const (
	messagePath = "/messages/"
	ContentType = "Content-Type"
	Accept      = "Accept"
)

// Function messageHandler is an http handler that sends the http body to the producer, replying
// immediately with a successful http response.
func (g *gateway) messagesHandler(w http.ResponseWriter, r *http.Request) {
	topic, err := parseTopic(r, messagePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = g.producer.Send(topic, message.NewMessage(b, propagateIncomingHeaders(r)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Message published to topic: %s\n", topic)
}

func propagateIncomingHeaders(request *http.Request) message.Headers {
	var defaultHeaders = []string{ContentType, Accept}
	var whitelistedHeaders = strings.Split(os.Getenv("RIFF_HTTP_HEADERS_WHITELIST"), ",")
	incomingHeadersToPropagate := append(defaultHeaders, whitelistedHeaders...)

	header := make(message.Headers)
	for _, h := range incomingHeadersToPropagate {
		if vs, ok := request.Header[h]; ok {
			header[h] = vs
		}
	}
	return header
}
