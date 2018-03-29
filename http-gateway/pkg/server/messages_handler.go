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
	"log"
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
	topicName, err := parseTopic(r, messagePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !g.topicHelper.TopicExists(topicName) {
		errMsg := fmt.Sprintf("could not find topic '%s'", topicName)
		log.Printf(errMsg)
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = g.producer.Send(topicName, message.NewMessage(b, propagateIncomingHeaders(r)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Message published to topic: %s\n", topicName)
}

func propagateIncomingHeaders(request *http.Request) message.Headers {
	var defaultHeaders = []string{ContentType, Accept}
	var whitelistedHeaders = strings.Split(os.Getenv("RIFF_HTTP_HEADERS_WHITELIST"), ",")
	incomingHeadersToPropagate := append(defaultHeaders, whitelistedHeaders...)

	header := make(message.Headers)
	for _, h := range incomingHeadersToPropagate {
		if vs, ok := request.Header[http.CanonicalHeaderKey(h)]; ok {
			header[http.CanonicalHeaderKey(h)] = vs
		}
	}
	return header
}
