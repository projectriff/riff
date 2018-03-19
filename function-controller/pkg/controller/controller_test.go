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
package controller_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/function-controller/mocks"
	"github.com/projectriff/riff/function-controller/pkg/controller"
	"github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler/mockautoscaler"
	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler"
)

var _ = Describe("Controller", func() {
	var (
		ctrl                controller.Controller
		deployer            *mocks.Deployer
		autoScaler          *mockautoscaler.AutoScaler
		deploymentsHandlers cache.ResourceEventHandlerFuncs
		functionHandlers    cache.ResourceEventHandlerFuncs
		topicHandlers       cache.ResourceEventHandlerFuncs
		closeCh             chan struct{}
		maxReplicasPolicy   func(id autoscaler.FunctionId) int
		delayScaleDownPolicy func(function autoscaler.FunctionId) time.Duration
	)

	BeforeEach(func() {
		deployer = new(mocks.Deployer)

		topicInformer := new(mocks.TopicInformer)
		functionInformer := new(mocks.FunctionInformer)
		deploymentInformer := new(mocks.DeploymentInformer)

		siiTopics := new(mocks.SharedIndexInformer)
		topicInformer.On("Informer").Return(siiTopics)
		siiTopics.On("AddEventHandler", mock.AnythingOfType("cache.ResourceEventHandlerFuncs")).Run(func(args mock.Arguments) {
			topicHandlers = args.Get(0).(cache.ResourceEventHandlerFuncs)
		})
		siiTopics.On("Run", mock.Anything)

		siiFunctions := new(mocks.SharedIndexInformer)
		functionInformer.On("Informer").Return(siiFunctions)
		siiFunctions.On("AddEventHandler", mock.AnythingOfType("cache.ResourceEventHandlerFuncs")).Run(func(args mock.Arguments) {
			functionHandlers = args.Get(0).(cache.ResourceEventHandlerFuncs)
		})
		siiFunctions.On("Run", mock.Anything)

		siiDeployments := new(mocks.SharedIndexInformer)
		deploymentInformer.On("Informer").Return(siiDeployments)
		siiDeployments.On("AddEventHandler", mock.AnythingOfType("cache.ResourceEventHandlerFuncs")).Run(func(args mock.Arguments) {
			deploymentsHandlers = args.Get(0).(cache.ResourceEventHandlerFuncs)
		})
		siiDeployments.On("Run", mock.Anything)

		autoScaler = new(mockautoscaler.AutoScaler)
		autoScaler.On("SetMaxReplicasPolicy", mock.AnythingOfType("func(autoscaler.FunctionId) int")).Run(func(args mock.Arguments) {
			maxReplicasPolicy = args.Get(0).(func(id autoscaler.FunctionId) int)
		})
		autoScaler.On("SetDelayScaleDownPolicy", mock.AnythingOfType("func(autoscaler.FunctionId) time.Duration")).Run(func(args mock.Arguments) {
			delayScaleDownPolicy = args.Get(0).(func(function autoscaler.FunctionId) time.Duration)
		})
		autoScaler.On("Run")
		autoScaler.On("Close").Return(nil)
		autoScaler.On("StartMonitoring", mock.AnythingOfType("string"), mock.AnythingOfType("autoscaler.FunctionId")).Return(nil)
		autoScaler.On("StopMonitoring", mock.AnythingOfType("string"), mock.AnythingOfType("autoscaler.FunctionId")).Return(nil)
		autoScaler.On("InformFunctionReplicas", mock.AnythingOfType("autoscaler.FunctionId"), mock.AnythingOfType("int"))

		ctrl = controller.New(topicInformer, functionInformer, deploymentInformer, deployer, autoScaler, -1)
		closeCh = make(chan struct{}, 2) // 2 allows to easily send in a .Runt() func() on stubs w/o blocking
	})

	AfterEach(func() {
		deployer.AssertExpectations(GinkgoT())
	})

	It("should shut down properly", func() {
		proposal := make(map[autoscaler.FunctionId]int)
		autoScaler.On("Propose").Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		}).Return(proposal)
		ctrl.Run(closeCh)
	})

	It("should handle functions coming and going", func() {
		ctrl.SetScalingInterval(10 * time.Millisecond)

		fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
		deployer.On("Deploy", fn).Return(nil)
		proposal := make(map[autoscaler.FunctionId]int)
		proposal[autoscaler.FunctionId{"fn"}] = 1
		autoScaler.On("Propose").Return(proposal)

		deployer.On("Scale", fn, 1).Return(nil).Run(func(args mock.Arguments) {
			functionHandlers.DeleteFunc(fn)
		})
		deployer.On("Undeploy", fn).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})
		functionHandlers.AddFunc(fn)

		ctrl.Run(closeCh)
	})

	It("should handle functions being updated", func() {
		fn1 := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
		fn2 := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input2"}}

		deployer.On("Update", fn2, 0).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})

		functionHandlers.UpdateFunc(fn1, fn2)

		ctrl.Run(closeCh)
	})

	It("should handle a non-trivial input topic", func() {
		ctrl.SetScalingInterval(10 * time.Millisecond)

		fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
		deployer.On("Deploy", fn).Return(nil)

		three := int32(3)
		topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &three}}
		deployer.On("Deploy", fn).Return(nil)

		proposal := make(map[autoscaler.FunctionId]int)
		proposal[autoscaler.FunctionId{"fn"}] = 1
		autoScaler.On("Propose").Return(proposal).Once()

		proposal = make(map[autoscaler.FunctionId]int)
		proposal[autoscaler.FunctionId{"fn"}] = 2
		autoScaler.On("Propose").Return(proposal).Once()

		proposal = make(map[autoscaler.FunctionId]int)
		proposal[autoscaler.FunctionId{"fn"}] = 3
		autoScaler.On("Propose").Return(proposal).Once()

		deployer.On("Scale", fn, 1).Return(nil)
		deployer.On("Scale", fn, 2).Return(nil)
		deployer.On("Scale", fn, 3).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})
		functionHandlers.AddFunc(fn)
		topicHandlers.AddFunc(topic)

		ctrl.Run(closeCh)
	})

	// Tests that when actual replicas are disrupted (for whatever reason), the controller eventually requests a scaling
	// to the number of replicas it thinks is correct.
	It("should reconcile replicas on disruption", func() {
		ctrl.SetScalingInterval(10 * time.Millisecond)

		fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
		deployer.On("Deploy", fn).Return(nil)

		computes := 0
		three := int32(3)
		topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &three}}
		deployer.On("Deploy", fn).Return(nil)

		proposal := make(map[autoscaler.FunctionId]int)
		proposal[autoscaler.FunctionId{"fn"}] = 2
		autoScaler.On("Propose").Return(proposal).Times(5).Run(func(args mock.Arguments) {
			computes++
		})

		autoScaler.On("Propose").Return(proposal).Once().Run(func(args mock.Arguments) {
			computes++
			// Disrupt actual replicas on 6th computation
			deployment := v1beta1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "fn",
					Labels: map[string]string{"function": "fn"},
				},
				Status: v1beta1.DeploymentStatus{Replicas: int32(6)},
			}
			deploymentsHandlers.UpdateFunc(&deployment, &deployment)
		})

		autoScaler.On("Propose").Return(proposal).Once().Run(func(args mock.Arguments) {
			computes++
		})

		deployer.On("Scale", fn, 2).Return(nil).Once()
		deployer.On("Scale", fn, 2).Return(nil).Run(func(args mock.Arguments) {
			Expect(computes).To(Equal(7))
			closeCh <- struct{}{}
		})
		functionHandlers.AddFunc(fn)
		topicHandlers.AddFunc(topic)

		ctrl.Run(closeCh)
	})

	Describe("maxReplicasScalingPolicy", func() {
		Context("when the input topic has 10 partitions", func() {
			BeforeEach(func() {
				ten := int32(10)
				topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &ten}}
				topicHandlers.AddFunc(topic)

				proposal := make(map[autoscaler.FunctionId]int)
				proposal[autoscaler.FunctionId{"fn"}] = 0
				autoScaler.On("Propose").Return(proposal)

				go ctrl.Run(closeCh)
			})

			Context("when the function does not specify maxReplicas", func() {
			    BeforeEach(func() {
					fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
					deployer.On("Deploy", fn).Return(nil)
					functionHandlers.AddFunc(fn)
				})

				It("should eventually return 10", func() {
					// The controller takes a little while to set up the topic and function.
					Eventually(func() int { return maxReplicasPolicy(autoscaler.FunctionId{"fn"}); }).Should(Equal(10))
					closeCh <- struct{}{}
				})
			})


			Context("when the function specifies maxReplicas as 5", func() {
				BeforeEach(func() {
					five := int32(5)
					fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input", MaxReplicas: &five}}
					deployer.On("Deploy", fn).Return(nil)
					functionHandlers.AddFunc(fn)
				})

				It("should eventually return 5", func() {
					// The controller takes a little while to update the function.
					Eventually(func() int { return maxReplicasPolicy(autoscaler.FunctionId{"fn"}); }).Should(Equal(5))
					closeCh <- struct{}{}
				})
			})
		})
	})

	Describe("delayScaleDownPolicy", func() {
		Context("when the function does not specify idleTimeoutMs", func() {
		    BeforeEach(func() {
				fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
				deployer.On("Deploy", fn).Return(nil)
				functionHandlers.AddFunc(fn)

				proposal := make(map[autoscaler.FunctionId]int)
				proposal[autoscaler.FunctionId{"fn"}] = 0
				autoScaler.On("Propose").Return(proposal)

				go ctrl.Run(closeCh)
			})

			It("should consistently return the default scale down delay", func() {
				// The controller takes a little while to set up the topic and function.
				Consistently(func() time.Duration { return delayScaleDownPolicy(autoscaler.FunctionId{"fn"}); }).Should(Equal(time.Second*10))
				closeCh <- struct{}{}
			})
		})

		Context("when the function specifies idleTimeoutMs", func() {
			var idleTimeoutMs int32
		    BeforeEach(func() {
				idleTimeoutMs = 300
				fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input", IdleTimeoutMs: &idleTimeoutMs}}
				deployer.On("Deploy", fn).Return(nil)
				functionHandlers.AddFunc(fn)

				proposal := make(map[autoscaler.FunctionId]int)
				proposal[autoscaler.FunctionId{"fn"}] = 0
				autoScaler.On("Propose").Return(proposal)

				go ctrl.Run(closeCh)
			})

			It("should eventually return the specified scale down delay", func() {
				// The controller takes a little while to set up the function.
				Eventually(func() time.Duration { return delayScaleDownPolicy(autoscaler.FunctionId{"fn"}); }).Should(Equal(time.Millisecond*time.Duration(idleTimeoutMs)))
				closeCh <- struct{}{}
			})
		})
	})
})
