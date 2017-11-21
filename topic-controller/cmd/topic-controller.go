/*
 * Copyright 2016-2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"time"

	"github.com/sk8sio/sk8s-types/pkg/apis/extensions.sk8s.io/v1"
	"github.com/sk8sio/sk8s-types/pkg/client/topic"
	"github.com/sk8sio/topic-controller/pkg/topic/provisioner/kafka"

	"flag"
	"fmt"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"context"
	"os/signal"
	"syscall"
)

// return rest config, if path not specified assume in cluster config
func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func healthHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"UP"}`))
	}
}

func startHttpServer() *http.Server {
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/health", healthHandler())

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	log.Printf("Listening on %v", srv.Addr)
	return srv
}

func main() {

	kubeconf := flag.String("kubeconf", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := getClientConfig(*kubeconf)
	if err != nil {
		panic(err.Error())
	}

	// create clientset and create the topics CRD, if not already there
	clientset, err := apiextcs.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// note: if the CRD exist our CreateCRD function is set to exit without an error
	err = topic.CreateCRD(clientset)
	if err != nil {
		panic(err)
	}

	// Create a new clientset which include our CRD schema
	crdcs, scheme, err := topic.NewClient(config)
	if err != nil {
		panic(err)
	}

	// Create a CRD client interface
	crdclient := topic.TopicClient(crdcs, scheme, "default")

	provisioner := kafka.NewKafkaProvisioner(os.Getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_ZK_NODES"))

	// Watch for changes in Topic objects and fire Add, Delete, Update callbacks
	_, controller := cache.NewInformer(
		crdclient.NewListWatch(),
		&v1.Topic{},
		time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				t := obj.(*v1.Topic)
				t = applyDefaults(t)
				log.Printf("Adding topic %v with %v partitions", t.Name, *t.Spec.Partitions)
				err := provisioner.ProvisionProducerDestination(t.Name, int(*t.Spec.Partitions))
				if err != nil {
					panic(err)
				}
			},
			DeleteFunc: func(obj interface{}) {
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	srv := startHttpServer()

	// Trap signals to trigger a proper shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, os.Kill)

	// Wait for shutdown
	select {

	case <-signals:
		fmt.Println("Shutting Down...")
		timeout, c := context.WithTimeout(context.Background(), 1*time.Second)
		defer c()
		if err := srv.Shutdown(timeout); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
	}
}

func applyDefaults(topic *v1.Topic) *v1.Topic {
	if topic.Spec.Partitions == nil {
		defaultPartitions := int32(1)
		topic.Spec.Partitions = &defaultPartitions
	}
	return topic
}
