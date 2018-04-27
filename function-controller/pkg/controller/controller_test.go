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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/function-controller/mocks"
	"github.com/projectriff/riff/function-controller/pkg/controller"
	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler"
	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler/mockautoscaler"
	v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

var _ = Describe("Controller", func() {
	var (
		ctrl                 controller.Controller
		deployer             *mocks.Deployer
		autoScaler           *mockautoscaler.AutoScaler
		linkHandlers         cache.ResourceEventHandlerFuncs
		deploymentsHandlers  cache.ResourceEventHandlerFuncs
		functionHandlers     cache.ResourceEventHandlerFuncs
		topicHandlers        cache.ResourceEventHandlerFuncs
		closeCh              chan struct{}
		maxReplicasPolicy    func(id autoscaler.LinkId) int
		delayScaleDownPolicy func(function autoscaler.LinkId) time.Duration
	)

	BeforeEach(func() {
		deployer = new(mocks.Deployer)

		topicInformer := new(mocks.TopicInformer)
		functionInformer := new(mocks.FunctionInformer)
		linkInformer := new(mocks.LinkInformer)
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

		siiLinks := new(mocks.SharedIndexInformer)
		linkInformer.On("Informer").Return(siiLinks)
		siiLinks.On("AddEventHandler", mock.AnythingOfType("cache.ResourceEventHandlerFuncs")).Run(func(args mock.Arguments) {
			linkHandlers = args.Get(0).(cache.ResourceEventHandlerFuncs)
		})
		siiLinks.On("Run", mock.Anything)

		siiDeployments := new(mocks.SharedIndexInformer)
		deploymentInformer.On("Informer").Return(siiDeployments)
		siiDeployments.On("AddEventHandler", mock.AnythingOfType("cache.ResourceEventHandlerFuncs")).Run(func(args mock.Arguments) {
			deploymentsHandlers = args.Get(0).(cache.ResourceEventHandlerFuncs)
		})
		siiDeployments.On("Run", mock.Anything)

		autoScaler = new(mockautoscaler.AutoScaler)
		autoScaler.On("SetMaxReplicasPolicy", mock.AnythingOfType("func(autoscaler.LinkId) int")).Run(func(args mock.Arguments) {
			maxReplicasPolicy = args.Get(0).(func(id autoscaler.LinkId) int)
		})
		autoScaler.On("SetDelayScaleDownPolicy", mock.AnythingOfType("func(autoscaler.LinkId) time.Duration")).Run(func(args mock.Arguments) {
			delayScaleDownPolicy = args.Get(0).(func(function autoscaler.LinkId) time.Duration)
		})
		autoScaler.On("Run")
		autoScaler.On("Close").Return(nil)
		autoScaler.On("StartMonitoring", mock.AnythingOfType("string"), mock.AnythingOfType("autoscaler.LinkId")).Return(nil)
		autoScaler.On("StopMonitoring", mock.AnythingOfType("string"), mock.AnythingOfType("autoscaler.LinkId")).Return(nil)
		autoScaler.On("InformFunctionReplicas", mock.AnythingOfType("autoscaler.LinkId"), mock.AnythingOfType("int"))

		ctrl = controller.New(topicInformer, functionInformer, linkInformer, deploymentInformer, deployer, autoScaler, -1)
		closeCh = make(chan struct{}, 2) // 2 allows to easily send in a .Runt() func() on stubs w/o blocking
	})

	AfterEach(func() {
		deployer.AssertExpectations(GinkgoT())
	})

	It("should shut down properly", func() {
		proposal := make(map[autoscaler.LinkId]int)
		autoScaler.On("Propose").Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		}).Return(proposal)
		ctrl.Run(closeCh)
	})

	It("should create, update and remove a link for a function", func() {
		function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}}
		link1 := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.LinkSpec{Function: "fn", Input: "input1"}}
		link2 := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.LinkSpec{Function: "fn", Input: "input2"}}

		deployer.On("Deploy", link1, function).Return(nil).Run(func(args mock.Arguments) {
			linkHandlers.UpdateFunc(link1, link2)
		})
		deployer.On("Update", link2, function, 0).Return(nil).Run(func(args mock.Arguments) {
			linkHandlers.DeleteFunc(link2)
		})
		deployer.On("Undeploy", link2).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})

		functionHandlers.AddFunc(function)
		linkHandlers.AddFunc(link1)

		ctrl.Run(closeCh)
	})

	It("should handle multiple links for a function", func() {
		function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}}
		link1 := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "fn1"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
		link2 := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "fn2"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}

		deployer.On("Deploy", link1, function).Return(nil).Run(func(args mock.Arguments) {
			linkHandlers.AddFunc(link2)
		})
		deployer.On("Deploy", link2, function).Return(nil).Run(func(args mock.Arguments) {
			linkHandlers.DeleteFunc(link1)
		})
		deployer.On("Undeploy", link1).Return(nil).Run(func(args mock.Arguments) {
			linkHandlers.DeleteFunc(link2)
		})
		deployer.On("Undeploy", link2).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})

		functionHandlers.AddFunc(function)
		linkHandlers.AddFunc(link1)

		ctrl.Run(closeCh)
	})

	It("should handle link coming and going", func() {
		ctrl.SetScalingInterval(10 * time.Millisecond)

		function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}}
		link := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
		deployer.On("Deploy", link, function).Return(nil)
		proposal := make(map[autoscaler.LinkId]int)
		proposal[autoscaler.LinkId{"fn"}] = 1
		autoScaler.On("Propose").Return(proposal)

		deployer.On("Scale", link, 1).Return(nil).Run(func(args mock.Arguments) {
			fmt.Println("Scale")
			linkHandlers.DeleteFunc(link)
		})
		deployer.On("Undeploy", link).Return(nil).Run(func(args mock.Arguments) {
			fmt.Println("Undeploy")
			closeCh <- struct{}{}
		})

		functionHandlers.AddFunc(function)
		linkHandlers.AddFunc(link)

		ctrl.Run(closeCh)
	})

	It("should handle links being updated", func() {
		fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}}
		link1 := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
		link2 := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input2"}}

		deployer.On("Deploy", link1, fn).Return(nil).Run(func(args mock.Arguments) {
			linkHandlers.UpdateFunc(link1, link2)
		})

		deployer.On("Update", link2, fn, 0).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})

		functionHandlers.AddFunc(fn)
		linkHandlers.AddFunc(link1)

		ctrl.Run(closeCh)
	})

	It("should handle functions being updated", func() {
		fn1 := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Protocol: "http"}}
		fn2 := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Protocol: "grpc"}}
		link := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}

		deployer.On("Deploy", link, fn1).Return(nil).Run(func(args mock.Arguments) {
			functionHandlers.UpdateFunc(fn1, fn2)
		})

		deployer.On("Update", link, fn2, 0).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})

		functionHandlers.AddFunc(fn1)
		linkHandlers.AddFunc(link)

		ctrl.Run(closeCh)
	})

	It("should handle a non-trivial input topic", func() {
		ctrl.SetScalingInterval(10 * time.Millisecond)

		function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}}
		link := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
		deployer.On("Deploy", link, function).Return(nil)

		three := int32(3)
		topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &three}}
		deployer.On("Deploy", link, function).Return(nil)

		proposal := make(map[autoscaler.LinkId]int)
		proposal[autoscaler.LinkId{"link"}] = 1
		autoScaler.On("Propose").Return(proposal).Once()

		proposal = make(map[autoscaler.LinkId]int)
		proposal[autoscaler.LinkId{"link"}] = 2
		autoScaler.On("Propose").Return(proposal).Once()

		proposal = make(map[autoscaler.LinkId]int)
		proposal[autoscaler.LinkId{"link"}] = 3
		autoScaler.On("Propose").Return(proposal).Once()

		deployer.On("Scale", link, 1).Return(nil)
		deployer.On("Scale", link, 2).Return(nil)
		deployer.On("Scale", link, 3).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})

		functionHandlers.AddFunc(function)
		topicHandlers.AddFunc(topic)
		linkHandlers.AddFunc(link)

		ctrl.Run(closeCh)
	})

	// Tests that when actual replicas are disrupted (for whatever reason), the controller eventually requests a scaling
	// to the number of replicas it thinks is correct.
	It("should reconcile replicas on disruption", func() {
		ctrl.SetScalingInterval(10 * time.Millisecond)

		function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}}
		link := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
		deployer.On("Deploy", link, function).Return(nil)

		computes := 0
		three := int32(3)
		topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &three}}
		deployer.On("Deploy", link, function).Return(nil)

		proposal := make(map[autoscaler.LinkId]int)
		proposal[autoscaler.LinkId{"link"}] = 2
		autoScaler.On("Propose").Return(proposal).Times(5).Run(func(args mock.Arguments) {
			computes++
		})

		autoScaler.On("Propose").Return(proposal).Once().Run(func(args mock.Arguments) {
			computes++
			// Disrupt actual replicas on 6th computation
			deployment := v1beta1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "link",
					Labels: map[string]string{"function": "fn", "link": "link"},
				},
				Status: v1beta1.DeploymentStatus{Replicas: int32(6)},
			}
			deploymentsHandlers.UpdateFunc(&deployment, &deployment)
		})

		autoScaler.On("Propose").Return(proposal).Once().Run(func(args mock.Arguments) {
			computes++
		})

		deployer.On("Scale", link, 2).Return(nil).Once()
		deployer.On("Scale", link, 2).Return(nil).Run(func(args mock.Arguments) {
			Expect(computes).To(Equal(7))
			closeCh <- struct{}{}
		})

		functionHandlers.AddFunc(function)
		topicHandlers.AddFunc(topic)
		linkHandlers.AddFunc(link)

		ctrl.Run(closeCh)
	})

	Describe("maxReplicasScalingPolicy", func() {
		Context("when the input topic has 10 partitions", func() {
			BeforeEach(func() {
				ten := int32(10)
				topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &ten}}
				topicHandlers.AddFunc(topic)

				proposal := make(map[autoscaler.LinkId]int)
				proposal[autoscaler.LinkId{"fn"}] = 0
				autoScaler.On("Propose").Return(proposal)

				go ctrl.Run(closeCh)
			})

			Context("when the function does not specify maxReplicas", func() {
				BeforeEach(func() {
					function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}}
					link := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
					deployer.On("Deploy", link, function).Return(nil)
					functionHandlers.AddFunc(function)
					linkHandlers.AddFunc(link)
				})

				It("should eventually return 10", func() {
					// The controller takes a little while to set up the topic and function.
					Eventually(func() int { return maxReplicasPolicy(autoscaler.LinkId{"link"}) }).Should(Equal(10))
					closeCh <- struct{}{}
				})
			})

			Context("when the function specifies maxReplicas as 5", func() {
				BeforeEach(func() {
					five := int32(5)
					function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{MaxReplicas: &five}}
					link := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
					deployer.On("Deploy", link, function).Return(nil)
					functionHandlers.AddFunc(function)
					linkHandlers.AddFunc(link)
				})

				It("should eventually return 5", func() {
					// The controller takes a little while to update the function.
					Eventually(func() int { return maxReplicasPolicy(autoscaler.LinkId{"link"}) }).Should(Equal(5))
					closeCh <- struct{}{}
				})
			})
		})
	})

	Describe("delayScaleDownPolicy", func() {
		Context("when the function does not specify idleTimeoutMs", func() {
			BeforeEach(func() {
				function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}}
				link := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
				deployer.On("Deploy", link, function).Return(nil)
				functionHandlers.AddFunc(function)
				linkHandlers.AddFunc(link)

				proposal := make(map[autoscaler.LinkId]int)
				proposal[autoscaler.LinkId{"link"}] = 0
				autoScaler.On("Propose").Return(proposal)

				go ctrl.Run(closeCh)
			})

			It("should consistently return the default scale down delay", func() {
				// The controller takes a little while to set up the topic and function.
				Consistently(func() time.Duration { return delayScaleDownPolicy(autoscaler.LinkId{"link"}) }).Should(Equal(time.Second * 10))
				closeCh <- struct{}{}
			})
		})

		Context("when the function specifies idleTimeoutMs", func() {
			var idleTimeoutMs int32
			BeforeEach(func() {
				idleTimeoutMs = 300
				function := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{IdleTimeoutMs: &idleTimeoutMs}}
				link := &v1.Link{ObjectMeta: metav1.ObjectMeta{Name: "link"}, Spec: v1.LinkSpec{Function: "fn", Input: "input"}}
				deployer.On("Deploy", link, function).Return(nil)
				functionHandlers.AddFunc(function)
				linkHandlers.AddFunc(link)

				proposal := make(map[autoscaler.LinkId]int)
				proposal[autoscaler.LinkId{"link"}] = 0
				autoScaler.On("Propose").Return(proposal)

				go ctrl.Run(closeCh)
			})

			It("should eventually return the specified scale down delay", func() {
				// The controller takes a little while to set up the function.
				Eventually(func() time.Duration { return delayScaleDownPolicy(autoscaler.LinkId{"link"}) }).Should(Equal(time.Millisecond * time.Duration(idleTimeoutMs)))
				closeCh <- struct{}{}
			})
		})
	})
})
