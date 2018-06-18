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

	riffcs "github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"
	informers "github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions"
	"github.com/projectriff/riff/topic-controller/pkg/topic/provisioner/kafka"

	"flag"
	"os"
	"os/signal"
	"syscall"

	"log"

	"github.com/golang/glog"
	informersV1 "github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/topic-controller/pkg/topic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// return rest config, if path not specified assume in cluster config
func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {
	kubeconf := flag.String("kubeconf", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	topicsInformer := makeTopicsInformer(kubeconf)

	provisioner := kafka.NewKafkaProvisioner(os.Getenv("KAFKA_ZK_NODES"))

	controller := topic.NewController(topicsInformer, provisioner, 8080)
	stop := make(chan struct{})
	controller.Run(stop)

	// Trap signals to trigger a proper shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, os.Kill)

	// Wait for shutdown
	select {
	case <-signals:
		log.Println("Shutting Down...")
		stop <- struct{}{}
	}
}
func makeTopicsInformer(kubeconf *string) informersV1.TopicInformer {
	config, err := getClientConfig(*kubeconf)
	if err != nil {
		glog.Fatalf("Error getting client config: %s", err.Error())
	}
	riffClient, err := riffcs.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Error building riff clientset: %s", err.Error())
	}
	riffInformerFactory := informers.NewSharedInformerFactory(riffClient, time.Second*30)
	topicsInformer := riffInformerFactory.Projectriff().V1alpha1().Topics()
	return topicsInformer
}
