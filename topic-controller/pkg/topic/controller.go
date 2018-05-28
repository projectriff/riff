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

package topic

import (
	"fmt"
	"log"
	"net/http"

	"context"
	"time"

	v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	informersV1 "github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/topic-controller/pkg/topic/provisioner"
	"k8s.io/client-go/tools/cache"
)

type Controller interface {
	Run(<-chan struct{})
}

// controller watches the 'topic' custom resource and creates new topics using the provisioner abstraction.
// it also embeds an http server implementing health probes.
type controller struct {
	topicsInformer informersV1.TopicInformer
	provisioner    provisioner.Provisioner
	httpServer     *http.Server
}

// Run starts the informer watching for topics changes, as well as an http server to answer health probes.
// The controller shuts down when a message is received on the passed in channel
func (c *controller) Run(stopCh <-chan struct{}) {

	// Run informer
	informerStop := make(chan struct{})
	go c.topicsInformer.Informer().Run(informerStop)

	// Run http server
	go func() {
		log.Printf("Listening on %v", c.httpServer.Addr)
		if err := c.httpServer.ListenAndServe(); err != nil {
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	// Wait for a close message on stopCh
	go func() {
		<-stopCh
		close(informerStop) // Shut down informer
		timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := c.httpServer.Shutdown(timeout); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
	}()

}

// NewController creates a ready to run controller, using the provided informer and provisioner and listening on the
// given http port
func NewController(topicsInformer informersV1.TopicInformer, provisioner provisioner.Provisioner, port int) Controller {

	ctrl := controller{topicsInformer: topicsInformer, provisioner: provisioner, httpServer: makeHttpServer(port)}

	// Set up an event handler for when topic resources are added
	ctrl.topicsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			t := obj.(*v1.Topic)
			t = applyDefaults(t)
			log.Printf("Adding topic %v with %v partitions", t.Name, *t.Spec.Partitions)
			err := provisioner.ProvisionProducerDestination(t.Name, int(*t.Spec.Partitions))
			if err != nil {
				log.Printf("Failed to add topic %v: %v", t.Name, err)
			}
		},
	})

	return &ctrl
}

func makeHttpServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"UP"}`))
	})

	return &http.Server{Addr: fmt.Sprintf(":%v", port), Handler: mux}
}

func applyDefaults(topic *v1.Topic) *v1.Topic {
	if topic.Spec.Partitions == nil {
		defaultPartitions := int32(1)
		topic.Spec.Partitions = &defaultPartitions
	}
	return topic
}
