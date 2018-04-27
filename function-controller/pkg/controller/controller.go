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

	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler"
	v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	informersV1 "github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions/projectriff/v1alpha1"
	"k8s.io/api/extensions/v1beta1"
	informersV1Beta1 "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

// DefaultScalerInterval controls how often to run the scaling strategy.
// 97ms is chosen to avoid accidental locksteps with other systems such
// as OS schedulers or garbage collection.
const DefaultScalerInterval = 97 * time.Millisecond

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
	functionsUpdated          chan *v1.Function
	functionsDeleted          chan *v1.Function
	bindingsAdded             chan *v1.Binding
	bindingsUpdated           chan deltaBind
	bindingsDeleted           chan *v1.Binding
	deploymentsAddedOrUpdated chan *v1beta1.Deployment // TODO investigate deprecation -> apps?
	deploymentsDeleted        chan *v1beta1.Deployment // TODO investigate deprecation -> apps?

	topicInformer      informersV1.TopicInformer
	functionInformer   informersV1.FunctionInformer
	bindingInformer    informersV1.BindingInformer
	deploymentInformer informersV1Beta1.DeploymentInformer

	functions      map[fnKey]*v1.Function
	topics         map[topicKey]*v1.Topic
	bindings       map[bindingKey]*v1.Binding
	actualReplicas map[bindingKey]int32

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

type bindingKey struct {
	name string
}

// A deltaBind represents a pair of bindings involved in an update
type deltaBind struct {
	before *v1.Binding
	after  *v1.Binding
}

// Run starts the main controller loop, which streamlines concurrent notifications of topics,
// functions, bindings and deployments coming and going, and periodically runs the function
// scaling logic.
func (c *ctrl) Run(stopCh <-chan struct{}) {

	// Run informer
	informerStop := make(chan struct{})
	go c.topicInformer.Informer().Run(informerStop)
	go c.functionInformer.Informer().Run(informerStop)
	go c.bindingInformer.Informer().Run(informerStop)
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
		case function := <-c.functionsUpdated:
			c.onFunctionUpdated(function)
		case function := <-c.functionsDeleted:
			c.onFunctionDeleted(function)
		case binding := <-c.bindingsAdded:
			c.onBindingAdded(binding)
		case deltaBind := <-c.bindingsUpdated:
			c.onBindingUpdated(deltaBind.before, deltaBind.after)
		case binding := <-c.bindingsDeleted:
			c.onBindingDeleted(binding)
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
	// TODO (maybe) update bindings for this topic
}

func (c *ctrl) onTopicDeleted(topic *v1.Topic) {
	log.Printf("Topic deleted: %v", topic.Name)
	delete(c.topics, tkey(topic))
}

func (c *ctrl) onFunctionAdded(function *v1.Function) {
	log.Printf("Function added: %v", function.Name)
	c.functions[fkey(function)] = function
	// create deployments for pre-existing bindings
	for _, binding := range c.collectBindings(function) {
		c.createDeployment(binding, function)
	}
}

func (c *ctrl) onFunctionUpdated(function *v1.Function) {
	log.Printf("Function updated: %v", function.Name)
	c.functions[fkey(function)] = function
	// trigger binding updates
	for _, binding := range c.collectBindings(function) {
		c.onBindingUpdated(binding, binding)
	}
}

func (c *ctrl) onFunctionDeleted(function *v1.Function) {
	log.Printf("Function deleted: %v", function.Name)
	delete(c.functions, fkey(function))
}

func (c *ctrl) onBindingAdded(binding *v1.Binding) {
	log.Printf("Binding added: %v", binding.Name)
	c.bindings[bkey(binding)] = binding
	function := c.functions[fnKey{binding.Spec.Handler}]
	if function != nil {
		c.createDeployment(binding, function)
	}
}

func (c *ctrl) onBindingUpdated(oldBind *v1.Binding, newBind *v1.Binding) {
	if oldBind.Name != newBind.Name {
		log.Printf("Error: binding name cannot change on update: %s -> %s", oldBind.Name, newBind.Name)
		return
	}
	if oldBind.Namespace != newBind.Namespace {
		log.Printf("Error: binding namespace cannot change on update: %s -> %s", oldBind.Namespace, newBind.Namespace)
		return
	}
	log.Printf("Binding updated: %v", oldBind.Name)

	bindKey := bkey(oldBind)
	c.bindings[bindKey] = newBind

	if newBind.Spec.Input != oldBind.Spec.Input {
		c.autoscaler.StopMonitoring(oldBind.Spec.Input, autoscaler.FunctionId{oldBind.Name})

		c.autoscaler.StartMonitoring(newBind.Spec.Input, autoscaler.FunctionId{newBind.Name})
	}

	function := c.functions[fnKey{newBind.Spec.Handler}]
	if function != nil {
		err := c.deployer.Update(newBind, function, int(c.actualReplicas[bindKey]))
		if err != nil {
			log.Printf("Error %v", err)
		}
	}
}

func (c *ctrl) onBindingDeleted(binding *v1.Binding) {
	log.Printf("Binding deleted: %v", binding.Name)
	delete(c.bindings, bkey(binding))
	err := c.deployer.Undeploy(binding)
	if err != nil {
		log.Printf("Error %v", err)
	}
	c.autoscaler.StopMonitoring(binding.Spec.Input, autoscaler.FunctionId{binding.Name})
}

func (c *ctrl) collectBindings(function *v1.Function) []*v1.Binding {
	matches := []*v1.Binding{}
	functionKey := fkey(function)
	for _, binding := range c.bindings {
		handlerKey := bhkey(binding)
		if functionKey == handlerKey {
			matches = append(matches, binding)
		}
	}
	return matches
}

func (c *ctrl) createDeployment(binding *v1.Binding, function *v1.Function) {
	// TODO create owner references
	err := c.deployer.Deploy(binding, function)
	if err != nil {
		log.Printf("Error %v", err)
	}
	// TODO maybe rename autoscaler.FunctionId to DeploymentId
	c.autoscaler.StartMonitoring(binding.Spec.Input, autoscaler.FunctionId{binding.Name})
}

func (c *ctrl) onDeploymentAddedOrUpdated(deployment *v1beta1.Deployment) {
	if key := bindKey(deployment); key != nil {
		log.Printf("Deployment added/updated: %v", deployment.Name)
		c.actualReplicas[*key] = deployment.Status.Replicas
		c.autoscaler.InformFunctionReplicas(bindingKeyToId(key), int(deployment.Status.Replicas))
	}
}

func (c *ctrl) onDeploymentDeleted(deployment *v1beta1.Deployment) {
	if key := bindKey(deployment); key != nil {
		log.Printf("Deployment deleted: %v", deployment.Name)
		delete(c.actualReplicas, *key)
		c.autoscaler.InformFunctionReplicas(bindingKeyToId(key), 0)
	}
}

func bindKey(deployment *v1beta1.Deployment) *bindingKey {
	if deployment.Labels["function"] != "" {
		return &bindingKey{deployment.Labels["function"]}
	} else {
		return nil
	}
}

// TODO: unify bindingKey and autoscaler.FunctionId so conversion is not necessary
func bindingKeyToId(key *bindingKey) autoscaler.FunctionId {
	return autoscaler.FunctionId{key.name}
}

func fkey(function *v1.Function) fnKey {
	return fnKey{name: function.Name}
}
func bhkey(binding *v1.Binding) fnKey {
	return fnKey{name: binding.Spec.Handler}
}

func tkey(topic *v1.Topic) topicKey {
	return topicKey{name: topic.Name}
}

func bkey(binding *v1.Binding) bindingKey {
	return bindingKey{name: binding.Name}
}

func (c *ctrl) scale() {
	replicas := c.autoscaler.Propose()

	//log.Printf("Offsets = %v, =>Replicas = %v", offsets, replicas)

	for k, binding := range c.bindings {
		bindKey := bkey(binding)
		bindId := bindingKeyToId(&bindKey)
		desired := replicas[bindId]

		//log.Printf("For %v, want %v currently have %v", fn.Name, desired, c.actualReplicas[k])

		if int32(desired) != c.actualReplicas[k] {
			err := c.deployer.Scale(binding, desired)
			if err != nil {
				log.Printf("Error %v", err)
			}
			c.actualReplicas[k] = int32(desired)                 // This may also be updated by deployments informer later.
			c.autoscaler.InformFunctionReplicas(bindId, desired) // This may also be updated by deployments informer later.
		}
	}
}

// New initialises a new function controller, adding event handlers to the provided informers.
func New(topicInformer informersV1.TopicInformer,
	functionInformer informersV1.FunctionInformer,
	bindingInformer informersV1.BindingInformer,
	deploymentInformer informersV1Beta1.DeploymentInformer,
	deployer Deployer,
	auto autoscaler.AutoScaler,
	port int) Controller {

	pctrl := &ctrl{
		topicsAddedOrUpdated:      make(chan *v1.Topic, 100),
		topicsDeleted:             make(chan *v1.Topic, 100),
		topicInformer:             topicInformer,
		functionsAdded:            make(chan *v1.Function, 100),
		functionsUpdated:          make(chan *v1.Function, 100),
		functionsDeleted:          make(chan *v1.Function, 100),
		functionInformer:          functionInformer,
		bindingsAdded:             make(chan *v1.Binding, 100),
		bindingsUpdated:           make(chan deltaBind, 100),
		bindingsDeleted:           make(chan *v1.Binding, 100),
		bindingInformer:           bindingInformer,
		deploymentsAddedOrUpdated: make(chan *v1beta1.Deployment, 100),
		deploymentsDeleted:        make(chan *v1beta1.Deployment, 100),
		deploymentInformer:        deploymentInformer,
		functions:                 make(map[fnKey]*v1.Function),
		topics:                    make(map[topicKey]*v1.Topic),
		bindings:                  make(map[bindingKey]*v1.Binding),
		actualReplicas:            make(map[bindingKey]int32),
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
			fn := new.(*v1.Function)
			v1.SetObjectDefaults_Function(fn)
			pctrl.functionsUpdated <- fn
		},
		DeleteFunc: func(obj interface{}) {
			fn := obj.(*v1.Function)
			v1.SetObjectDefaults_Function(fn)
			pctrl.functionsDeleted <- fn
		},
	})
	bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			binding := obj.(*v1.Binding)
			v1.SetObjectDefaults_Binding(binding)
			pctrl.bindingsAdded <- binding
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldBind := old.(*v1.Binding)
			v1.SetObjectDefaults_Binding(oldBind)

			newBind := new.(*v1.Binding)
			v1.SetObjectDefaults_Binding(newBind)

			pctrl.bindingsUpdated <- deltaBind{before: oldBind, after: newBind}
		},
		DeleteFunc: func(obj interface{}) {
			binding := obj.(*v1.Binding)
			v1.SetObjectDefaults_Binding(binding)
			pctrl.bindingsDeleted <- binding
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

	auto.SetMaxReplicasPolicy(func(deployment autoscaler.FunctionId) int {
		replicas := 1
		if binding, ok := pctrl.bindings[bindingKey{deployment.Function}]; ok {
			if binding.Spec.Input != "" {
				if topic, ok := pctrl.topics[topicKey{binding.Spec.Input}]; ok {
					replicas = int(*topic.Spec.Partitions)
				}
			}

			if fn, ok := pctrl.functions[fnKey{binding.Spec.Handler}]; ok {
				if fn.Spec.MaxReplicas != nil {
					replicas = min(int(*fn.Spec.MaxReplicas), replicas)
				}
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
