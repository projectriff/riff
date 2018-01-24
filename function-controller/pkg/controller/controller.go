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
	"log"

	"time"

	"github.com/projectriff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	informersV1 "github.com/projectriff/kubernetes-crds/pkg/client/informers/externalversions/projectriff/v1"
	"k8s.io/api/extensions/v1beta1"
	informersV1Beta1 "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

// DefaultScalerInterval controls how often to run the scaling strategy.
const DefaultScalerInterval = 100 * time.Millisecond

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
type replicaCounts map[fnKey]int

// type activityCounts is a mapping from function to combined activity marker (we're using sum of position accross all
// partitions and all topics
type activityCounts map[fnKey]struct {
	current int64
	end     int64
}

// a scalingPolicy turns offsets into replicaCounts. activityCounts are computed as a byproduct
type scalingPolicy func(offsets map[Subscription]PartitionedOffsets) (replicaCounts, activityCounts)

// a scalingPostProcessor applies transformation on the computed number of replicas, possibly using activityCounts,
// which is expected to be passed along without change
type scalingPostProcessor func(replicaCounts, activityCounts) (replicaCounts, activityCounts)

type ctrl struct {
	topicsAddedOrUpdated      chan *v1.Topic
	topicsDeleted             chan *v1.Topic
	functionsAddedOrUpdated   chan *v1.Function
	functionsDeleted          chan *v1.Function
	deploymentsAddedOrUpdated chan *v1beta1.Deployment // TODO investigate deprecation -> apps?
	deploymentsDeleted        chan *v1beta1.Deployment // TODO investigate deprecation -> apps?

	topicInformer      informersV1.TopicInformer
	functionInformer   informersV1.FunctionInformer
	deploymentInformer informersV1Beta1.DeploymentInformer

	functions      map[fnKey]*v1.Function
	topics         map[topicKey]*v1.Topic
	actualReplicas map[fnKey]int

	deployer   Deployer
	lagTracker LagTracker

	scalerInterval time.Duration
	scalingPolicy  scalingPolicy
}

type fnKey struct {
	name string
	// TODO should include namespace as well
}

type topicKey struct {
	name string
}

// Run starts the main controller loop, which streamlines concurrent notifications of topics, functions and deployments
// coming and going, and periodically runs the function scaling logic.
func (c *ctrl) Run(stopCh <-chan struct{}) {

	// Run informer
	informerStop := make(chan struct{})
	go c.topicInformer.Informer().Run(informerStop)
	go c.functionInformer.Informer().Run(informerStop)
	go c.deploymentInformer.Informer().Run(informerStop)

	for {
		select {
		case topic := <-c.topicsAddedOrUpdated:
			c.onTopicAddedOrUpdated(topic)
		case topic := <-c.topicsDeleted:
			c.onTopicDeleted(topic)
		case function := <-c.functionsAddedOrUpdated:
			c.onFunctionAddedOrUpdated(function)
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
			return
		}
	}
}

func (c *ctrl) SetScalingInterval(interval time.Duration) {
	c.scalerInterval = interval
}

func (c *ctrl) SetScalingPolicy(policy scalingPolicy) {
	c.scalingPolicy = policy
}

func (c *ctrl) onTopicAddedOrUpdated(topic *v1.Topic) {
	log.Printf("Topic added: %v", topic.Name)
	c.topics[tkey(topic)] = topic
}

func (c *ctrl) onTopicDeleted(topic *v1.Topic) {
	log.Printf("Topic deleted: %v", topic.Name)
	delete(c.topics, tkey(topic))
}

func (c *ctrl) onFunctionAddedOrUpdated(function *v1.Function) {
	log.Printf("Function added: %v", function.Name)
	c.functions[key(function)] = function
	c.lagTracker.BeginTracking(Subscription{Topic: function.Spec.Input, Group: function.Name})
	err := c.deployer.Deploy(function)
	if err != nil {
		log.Printf("Error %v", err)
	}
}

func (c *ctrl) onFunctionDeleted(function *v1.Function) {
	log.Printf("Function deleted: %v", function.Name)
	delete(c.functions, key(function))
	c.lagTracker.StopTracking(Subscription{Topic: function.Spec.Input, Group: function.Name})
	err := c.deployer.Undeploy(function)
	if err != nil {
		log.Printf("Error %v", err)
	}
}

func (c *ctrl) onDeploymentAddedOrUpdated(deployment *v1beta1.Deployment) {
	if key := functionKey(deployment); key != nil {
		log.Printf("Deployment added/updated: %v", deployment.Name)
		c.actualReplicas[*key] = int(deployment.Status.Replicas)
	}
}

func (c *ctrl) onDeploymentDeleted(deployment *v1beta1.Deployment) {
	if key := functionKey(deployment); key != nil {
		log.Printf("Deployment deleted: %v", deployment.Name)
		delete(c.actualReplicas, *key)
	}
}

func functionKey(deployment *v1beta1.Deployment) *fnKey {
	if deployment.Labels["function"] != "" {
		return &fnKey{deployment.Labels["function"]}
	} else {
		return nil
	}
}

func key(function *v1.Function) fnKey {
	return fnKey{name: function.Name}
}

func tkey(topic *v1.Topic) topicKey {
	return topicKey{name: topic.Name}
}

func (c *ctrl) scale() {
	offsets := c.lagTracker.Compute()
	replicas, _ := c.scalingPolicy(offsets)

	//log.Printf("Offsets = %v, =>Replicas = %v", offsets, replicas)

	for k, fn := range c.functions {
		desired := replicas[key(fn)]

		//log.Printf("For %v, want %v currently have %v", fn.Name, desired, c.actualReplicas[k])

		if desired != c.actualReplicas[k] {
			err := c.deployer.Scale(fn, desired)
			c.actualReplicas[k] = desired // This may also be updated by deployments informer later.
			if err != nil {
				log.Printf("Error %v", err)
			}
		}
	}
}

// compose returns a function that applies 'post' after 'policy'
func compose(post scalingPostProcessor, policy scalingPolicy) scalingPolicy {
	return func(m map[Subscription]PartitionedOffsets) (replicaCounts, activityCounts) {
		return post(policy(m))
	}
}

// New initialises a new function controller, adding event handlers to the provided informers.
func New(topicInformer informersV1.TopicInformer,
	functionInformer informersV1.FunctionInformer,
	deploymentInformer informersV1Beta1.DeploymentInformer,
	deployer Deployer,
	tracker LagTracker) Controller {

	pctrl := &ctrl{
		topicsAddedOrUpdated:      make(chan *v1.Topic, 100),
		topicsDeleted:             make(chan *v1.Topic, 100),
		topicInformer:             topicInformer,
		functionsAddedOrUpdated:   make(chan *v1.Function, 100),
		functionsDeleted:          make(chan *v1.Function, 100),
		functionInformer:          functionInformer,
		deploymentsAddedOrUpdated: make(chan *v1beta1.Deployment, 100),
		deploymentsDeleted:        make(chan *v1beta1.Deployment, 100),
		deploymentInformer:        deploymentInformer,
		functions:                 make(map[fnKey]*v1.Function),
		topics:                    make(map[topicKey]*v1.Topic),
		actualReplicas:            make(map[fnKey]int),
		deployer:                  deployer,
		lagTracker:                tracker,
		scalerInterval:            DefaultScalerInterval,
	}
	pctrl.scalingPolicy = pctrl.computeDesiredReplicas

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
			pctrl.functionsAddedOrUpdated <- fn
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			fn := new.(*v1.Function)
			v1.SetObjectDefaults_Function(fn)
			pctrl.functionsAddedOrUpdated <- fn
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
	return pctrl
}

func DecorateWithDelayAndSmoothing(c Controller) {
	pctrl := c.(*ctrl)
	smoother := smoother{ctrl: pctrl, memory: make(map[fnKey]float32)}
	pctrl.scalingPolicy = compose(smoother.smooth, pctrl.scalingPolicy)

	delayer := delayer{ctrl: pctrl, scaleDownToZeroDecision: make(map[fnKey]time.Time), previousCombinedPositions: make(activityCounts)}
	pctrl.scalingPolicy = compose(delayer.delay, pctrl.scalingPolicy)
}

// computeDesiredReplicas turns a subscription based map (of offsets) into a function-key based map of how many replicas to spawn.
// The logic is as follows: for a given topic, look at how many partitions have lag or activity (there is no point in spawning 2
// replicas if only a single partition has lag, even if it's a 1000 messages lag).
// If the function has multiple inputs, we take the max of those computations (some topics may be starved but that's ok
// while we still maximize the throughput for the given topic that has the most needy partitions)
func (c *ctrl) computeDesiredReplicas(offsets map[Subscription]PartitionedOffsets) (replicaCounts, activityCounts) {
	replicas := make(replicaCounts)
	combinedCurrentPositions := make(activityCounts)
	for s, o := range offsets {
		fn := fnKey{s.Group}
		topic := topicKey{s.Topic}
		replicas[fn] = max(replicas[fn], c.desiredReplicasForTopic(o, fn, topic))
		ccp := combinedCurrentPositions[fn]
		ccp.end += sumEndPositions(o)
		ccp.current += sumCurrentPositions(o)
		combinedCurrentPositions[fn] = ccp
	}
	return replicas, combinedCurrentPositions
}

func sumEndPositions(offsets PartitionedOffsets) int64 {
	result := int64(0)
	for _, o := range offsets {
		result += o.End
	}
	return result
}

func sumCurrentPositions(offsets PartitionedOffsets) int64 {
	result := int64(0)
	for _, o := range offsets {
		result += o.Current
	}
	return result
}

func (c *ctrl) desiredReplicasForTopic(offsets PartitionedOffsets, fnKey fnKey, tKey topicKey) int {
	maxPartLag := int64(0)
	for _, o := range offsets {
		maxPartLag = max64(maxPartLag, o.Lag())
	}

	// TODO: those 3 numbers part of Function spec?
	lagRequiredForMax := float32(10)
	lagRequiredForOne := float32(1)
	minReplicas := int32(0)

	partitionCount := int32(1)
	if topic, ok := c.topics[tKey]; ok {
		partitionCount = *topic.Spec.Partitions
	}
	maxReplicas := partitionCount
	if fn, ok := c.functions[fnKey]; ok {
		if fn.Spec.MaxReplicas != nil {
			maxReplicas = *fn.Spec.MaxReplicas
		}
	}
	maxReplicas = clamp(maxReplicas, minReplicas, partitionCount)

	slope := (float32(maxReplicas) - 1.0) / (lagRequiredForMax - lagRequiredForOne)
	var computedReplicas int32
	if slope > 0.0 {
		// max>1
		computedReplicas = (int32)(1 + (float32(maxPartLag)-lagRequiredForOne)*slope)
	} else if maxPartLag >= int64(lagRequiredForOne) {
		computedReplicas = 1
	} else {
		computedReplicas = 0
	}
	return int(clamp(computedReplicas, minReplicas, maxReplicas))
}

func clamp(value int32, min int32, max int32) int32 {
	if value < min {
		value = min
	} else if value > max {
		value = max
	}
	return value
}

func max(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func max64(a int64, b int64) int64 {
	if a > b {
		return a
	} else {
		return b
	}
}
