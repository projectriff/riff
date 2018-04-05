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
	"time"

	"github.com/projectriff/riff/message-transport/pkg/message"

	"github.com/satori/go.uuid"
)

const (
	CorrelationId = "correlationId"
	Error         = "error"
	requestPath   = "/requests/"
)

var outgoingHeadersToPropagate = [...]string{ContentType}

// Function requestHandler is an http handler that sends the http body to the producer, then waits
// for a message on a go channel it creates for a reply and sends that as an http response.
func (g *gateway) requestsHandler(w http.ResponseWriter, r *http.Request) {
	topicName, err := parseTopic(r, requestPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	topicExists, err := g.riffTopicExistenceChecker.TopicExists(defaultNamespace, topicName)
	if err != nil {
		errMsg := fmt.Sprintf("while checking to see if there was a Riff topic '%s', an unexpected error occurred: %+v", topicName, err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	if !topicExists {
		errMsg := fmt.Sprintf("could not find Riff topic '%s'", topicName)
		http.Error(w, errMsg, http.StatusNotFound)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	correlationId := uuid.NewV4().String() // entropy bottleneck?
	replyChan := make(chan message.Message)
	g.replies.Put(correlationId, replyChan)
	defer g.replies.Delete(correlationId)

	headers := propagateIncomingHeaders(r)
	headers[CorrelationId] = []string{correlationId}

	err = g.producer.Send(topicName, message.NewMessage(b, headers))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	select {
	case reply := <-replyChan:
		replyError := reply.Headers().GetOrDefault(Error, "")
		if len(replyError) != 0 {
			// message is an error
			// TODO set status code based on replyError type
			w.WriteHeader(http.StatusInternalServerError)
		}
		propagateOutgoingHeaders(reply, w)
		w.Write(reply.Payload())
	case <-time.After(g.timeout):
		w.WriteHeader(404)
	}
}

func propagateOutgoingHeaders(message message.Message, response http.ResponseWriter) {
	for _, h := range outgoingHeadersToPropagate {
		if vs, ok := message.Headers()[h]; ok {
			response.Header()[h] = vs
		}
	}
}
