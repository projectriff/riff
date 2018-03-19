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

package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	informersV1 "github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions/projectriff/v1"
	"k8s.io/api/extensions/v1beta1"
	informersV1Beta1 "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler"
)

// DefaultScalerInterval controls how often to run the scaling strategy.
const DefaultScalerInterval = 100 * time.Millisecond

const defaultScaleDownDelay = time.Second * 10

// Controller deploys functions by monitoring input lag to registered functions. To do so, it periodically runs
// some scaling logic and keeps track of (un-)registered functions, topics and deployments.
type Controller interface {
	// Run requests that this controller starts doing its job, until an empty struct is sent on the close channel.
	Run(closeCh <-chan struct{})
	// SetScalerInterval changes the interval at which the controller recomputes the required number of replicas for functions.
	// Should not be called once running.
	SetScalingInterval(interval time.Duration)
}

// type replicaCounts is a mapping from function to wanted number of replicas
type replicaCounts map[fnKey]int32

// type activityCounts is a mapping from function to combined activity marker (we're using sum of position accross all
// partitions and all topics
type activityCounts map[fnKey]struct {
	current int64
	end     int64
}

type ctrl struct {
	topicsAddedOrUpdated      chan *v1.Topic
	topicsDeleted             chan *v1.Topic
	functionsAdded            chan *v1.Function
	functionsUpdated          chan deltaFn
	functionsDeleted          chan *v1.Function
	deploymentsAddedOrUpdated chan *v1beta1.Deployment // TODO investigate deprecation -> apps?
	deploymentsDeleted        chan *v1beta1.Deployment // TODO investigate deprecation -> apps?

	topicInformer      informersV1.TopicInformer
	functionInformer   informersV1.FunctionInformer
	deploymentInformer informersV1Beta1.DeploymentInformer

	functions      map[fnKey]*v1.Function
	topics         map[topicKey]*v1.Topic
	actualReplicas map[fnKey]int32

	autoscaler autoscaler.AutoScaler

	deployer Deployer

	scalerInterval time.Duration

	httpServer *http.Server
}

type fnKey struct {
	name string
	// TODO should include namespace as well
}

type topicKey struct {
	name string
}

// A deltaFn represents a pair of functions involved in an update
type deltaFn struct {
	before *v1.Function
	after  *v1.Function
}

// Run starts the main controller loop, which streamlines concurrent notifications of topics, functions and deployments
// coming and going, and periodically runs the function scaling logic.
func (c *ctrl) Run(stopCh <-chan struct{}) {

	// Run informer
	informerStop := make(chan struct{})
	go c.topicInformer.Informer().Run(informerStop)
	go c.functionInformer.Informer().Run(informerStop)
	go c.deploymentInformer.Informer().Run(informerStop)

	// Run autoscaler
	c.autoscaler.Run()

	for {
		select {
		case topic := <-c.topicsAddedOrUpdated:
			c.onTopicAddedOrUpdated(topic)
		case topic := <-c.topicsDeleted:
			c.onTopicDeleted(topic)
		case function := <-c.functionsAdded:
			c.onFunctionAdded(function)
		case deltaFn := <-c.functionsUpdated:
			c.onFunctionUpdated(deltaFn.before, deltaFn.after)
		case function := <-c.functionsDeleted:
			c.onFunctionDeleted(function)
		case deployment := <-c.deploymentsAddedOrUpdated:
			c.onDeploymentAddedOrUpdated(deployment)
		case deployment := <-c.deploymentsDeleted:
			c.onDeploymentDeleted(deployment)
		case <-time.After(c.scalerInterval):
			c.scale()
		case <-stopCh: // Maybe listen in another goroutine
			close(informerStop)
			c.autoscaler.Close()
			if c.httpServer != nil {
				timeout, ctx := context.WithTimeout(context.Background(), 1*time.Second)
				defer ctx()
				if err := c.httpServer.Shutdown(timeout); err != nil {
					panic(err) // failure/timeout shutting down the server gracefully
				}
			}
			return
		}
	}
}

func (c *ctrl) SetScalingInterval(interval time.Duration) {
	c.scalerInterval = interval
}

func (c *ctrl) onTopicAddedOrUpdated(topic *v1.Topic) {
	log.Printf("Topic added: %v", topic.Name)
	c.topics[tkey(topic)] = topic
}

func (c *ctrl) onTopicDeleted(topic *v1.Topic) {
	log.Printf("Topic deleted: %v", topic.Name)
	delete(c.topics, tkey(topic))
}

func (c *ctrl) onFunctionAdded(function *v1.Function) {
	log.Printf("Function added: %v", function.Name)
	c.functions[key(function)] = function
	err := c.deployer.Deploy(function)
	if err != nil {
		log.Printf("Error %v", err)
	}
	c.autoscaler.StartMonitoring(function.Spec.Input, autoscaler.FunctionId{function.Name})
}

func (c *ctrl) onFunctionUpdated(oldFn *v1.Function, newFn *v1.Function) {
	if oldFn.Name != newFn.Name {
		log.Printf("Error: function name cannot change on update: %s -> %s", oldFn.Name, newFn.Name)
		return
	}
	if oldFn.Namespace != newFn.Namespace {
		log.Printf("Error: function namespace cannot change on update: %s -> %s", oldFn.Namespace, newFn.Namespace)
		return
	}
	log.Printf("Function updated: %v", oldFn.Name)

	fnKey := key(oldFn)
	c.functions[fnKey] = newFn

	if newFn.Spec.Input != oldFn.Spec.Input {
		c.autoscaler.StopMonitoring(oldFn.Spec.Input, autoscaler.FunctionId{oldFn.Name})

		c.autoscaler.StartMonitoring(newFn.Spec.Input, autoscaler.FunctionId{newFn.Name})
	}

	err := c.deployer.Update(newFn, int(c.actualReplicas[fnKey]))
	if err != nil {
		log.Printf("Error %v", err)
	}
}

func (c *ctrl) onFunctionDeleted(function *v1.Function) {
	log.Printf("Function deleted: %v", function.Name)
	delete(c.functions, key(function))
	err := c.deployer.Undeploy(function)
	if err != nil {
		log.Printf("Error %v", err)
	}
	c.autoscaler.StopMonitoring(function.Spec.Input, autoscaler.FunctionId{function.Name})
}

func (c *ctrl) onDeploymentAddedOrUpdated(deployment *v1beta1.Deployment) {
	if key := functionKey(deployment); key != nil {
		log.Printf("Deployment added/updated: %v", deployment.Name)
		c.actualReplicas[*key] = deployment.Status.Replicas
		c.autoscaler.InformFunctionReplicas(fnKeyToId(key), int(deployment.Status.Replicas))
	}
}

func (c *ctrl) onDeploymentDeleted(deployment *v1beta1.Deployment) {
	if key := functionKey(deployment); key != nil {
		log.Printf("Deployment deleted: %v", deployment.Name)
		delete(c.actualReplicas, *key)
		c.autoscaler.InformFunctionReplicas(fnKeyToId(key), 0)
	}
}

func functionKey(deployment *v1beta1.Deployment) *fnKey {
	if deployment.Labels["function"] != "" {
		return &fnKey{deployment.Labels["function"]}
	} else {
		return nil
	}
}

// TODO: unify fnKey and autoscaler.FunctionId so conversion is not necessary
func fnKeyToId(key *fnKey) autoscaler.FunctionId {
	return autoscaler.FunctionId{key.name}
}

func key(function *v1.Function) fnKey {
	return fnKey{name: function.Name}
}

func tkey(topic *v1.Topic) topicKey {
	return topicKey{name: topic.Name}
}

func (c *ctrl) scale() {
	replicas := c.autoscaler.Propose()

	//log.Printf("Offsets = %v, =>Replicas = %v", offsets, replicas)

	for k, fn := range c.functions {
		fnKey := key(fn)
		fnId := fnKeyToId(&fnKey)
		desired := replicas[fnId]

		//log.Printf("For %v, want %v currently have %v", fn.Name, desired, c.actualReplicas[k])

		if int32(desired) != c.actualReplicas[k] {
			err := c.deployer.Scale(fn, desired)
			if err != nil {
				log.Printf("Error %v", err)
			}
			c.actualReplicas[k] = int32(desired)               // This may also be updated by deployments informer later.
			c.autoscaler.InformFunctionReplicas(fnId, desired) // This may also be updated by deployments informer later.
		}
	}
}

// New initialises a new function controller, adding event handlers to the provided informers.
func New(topicInformer informersV1.TopicInformer,
	functionInformer informersV1.FunctionInformer,
	deploymentInformer informersV1Beta1.DeploymentInformer,
	deployer Deployer,
	auto autoscaler.AutoScaler,
	port int) Controller {

	pctrl := &ctrl{
		topicsAddedOrUpdated:      make(chan *v1.Topic, 100),
		topicsDeleted:             make(chan *v1.Topic, 100),
		topicInformer:             topicInformer,
		functionsAdded:            make(chan *v1.Function, 100),
		functionsUpdated:          make(chan deltaFn, 100),
		functionsDeleted:          make(chan *v1.Function, 100),
		functionInformer:          functionInformer,
		deploymentsAddedOrUpdated: make(chan *v1beta1.Deployment, 100),
		deploymentsDeleted:        make(chan *v1beta1.Deployment, 100),
		deploymentInformer:        deploymentInformer,
		functions:                 make(map[fnKey]*v1.Function),
		topics:                    make(map[topicKey]*v1.Topic),
		actualReplicas:            make(map[fnKey]int32),
		deployer:                  deployer,
		autoscaler:                auto,
		scalerInterval:            DefaultScalerInterval,
	}
	topicInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			topic := obj.(*v1.Topic)
			v1.SetObjectDefaults_Topic(topic)
			pctrl.topicsAddedOrUpdated <- topic
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			topic := new.(*v1.Topic)
			v1.SetObjectDefaults_Topic(topic)
			pctrl.topicsAddedOrUpdated <- topic
		},
		DeleteFunc: func(obj interface{}) {
			topic := obj.(*v1.Topic)
			v1.SetObjectDefaults_Topic(topic)
			pctrl.topicsDeleted <- topic
		},
	})
	functionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fn := obj.(*v1.Function)
			v1.SetObjectDefaults_Function(fn)
			pctrl.functionsAdded <- fn
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldFn := old.(*v1.Function)
			v1.SetObjectDefaults_Function(oldFn)

			newFn := new.(*v1.Function)
			v1.SetObjectDefaults_Function(newFn)

			pctrl.functionsUpdated <- deltaFn{before: oldFn, after: newFn}
		},
		DeleteFunc: func(obj interface{}) {
			fn := obj.(*v1.Function)
			v1.SetObjectDefaults_Function(fn)
			pctrl.functionsDeleted <- fn
		},
	})
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { pctrl.deploymentsAddedOrUpdated <- obj.(*v1beta1.Deployment) },
		UpdateFunc: func(old interface{}, new interface{}) { pctrl.deploymentsAddedOrUpdated <- new.(*v1beta1.Deployment) },
		DeleteFunc: func(obj interface{}) { pctrl.deploymentsDeleted <- obj.(*v1beta1.Deployment) },
	})

	if port > 0 {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte(`{"status":"UP"}`))
		})
		addr := fmt.Sprintf(":%v", port)
		pctrl.httpServer = &http.Server{Addr: addr,
			Handler: mux,
		}
		go func() {
			log.Printf("Listening on %v", addr)
			if err := pctrl.httpServer.ListenAndServe(); err != nil {
				log.Printf("Httpserver: ListenAndServe() error: %s", err)
			}
		}()
	}

	auto.SetMaxReplicasPolicy(func(function autoscaler.FunctionId) int {
		replicas := 1
		if fn, ok := pctrl.functions[fnKey{function.Function}]; ok {
			if fn.Spec.Input != "" {
				if topic, ok := pctrl.topics[topicKey{fn.Spec.Input}]; ok {
					replicas = int(*topic.Spec.Partitions)
				}
			}

			if fn.Spec.MaxReplicas != nil {
				replicas = min(int(*fn.Spec.MaxReplicas), replicas)
			}
		}
		return replicas
	})

	auto.SetDelayScaleDownPolicy(func(function autoscaler.FunctionId) time.Duration {
		delay := defaultScaleDownDelay
		if fn, ok := pctrl.functions[fnKey{function.Function}]; ok {
			if fn.Spec.IdleTimeoutMs != nil {
				delay = time.Millisecond * time.Duration(*fn.Spec.IdleTimeoutMs)
			}
		}
		log.Printf("Delaying scaling down %v to 0 by %v", function, delay)
		return delay
	})

	return pctrl
}

func min(a int, b int) int {
	if a > b {
		return b
	}
	return a
}
