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
	"github.com/projectriff/function-controller/mocks"
	"github.com/projectriff/function-controller/pkg/controller"
	"github.com/projectriff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	"github.com/stretchr/testify/mock"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

var _ = Describe("Controller", func() {
	var (
		ctrl                controller.Controller
		tracker             *mocks.LagTracker
		deployer            *mocks.Deployer
		deploymentsHandlers cache.ResourceEventHandlerFuncs
		functionHandlers    cache.ResourceEventHandlerFuncs
		topicHandlers       cache.ResourceEventHandlerFuncs
		closeCh             chan struct{}
	)

	BeforeEach(func() {
		tracker = new(mocks.LagTracker)

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

		ctrl = controller.New(topicInformer, functionInformer, deploymentInformer, deployer, tracker, -1)
		closeCh = make(chan struct{}, 2) // 2 allows to easily send in a .Runt() func() on stubs w/o blocking
	})

	AfterEach(func() {
		tracker.AssertExpectations(GinkgoT())
		deployer.AssertExpectations(GinkgoT())
	})

	It("should shut down properly", func() {
		tracker.On("Compute").Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})
		ctrl.Run(closeCh)
	})

	It("should handle functions coming and going", func() {
		ctrl.SetScalingInterval(10 * time.Millisecond)

		tracker.On("BeginTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
		tracker.On("StopTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
		fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
		deployer.On("Deploy", fn).Return(nil)
		tracker.On("Compute").Return(lag(fn, 1))
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

		tracker.On("StopTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
		deployer.On("Update", fn2, 0).Return(nil).Run(func(args mock.Arguments) {
			closeCh <- struct{}{}
		})
		tracker.On("BeginTracking", controller.Subscription{Topic: "input2", Group: "fn"}).Return(nil)

		functionHandlers.UpdateFunc(fn1, fn2)

		ctrl.Run(closeCh)
	})

	It("should handle a non-trivial input topic", func() {
		ctrl.SetScalingInterval(10 * time.Millisecond)

		tracker.On("BeginTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
		fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
		deployer.On("Deploy", fn).Return(nil)

		three := int32(3)
		topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &three}}
		deployer.On("Deploy", fn).Return(nil)

		tracker.On("Compute").Return(lag(fn, 1, 0, 0)).Once()
		tracker.On("Compute").Return(lag(fn, 6, 0, 1)).Once()
		tracker.On("Compute").Return(lag(fn, 2, 3, 10)).Once()
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

		tracker.On("BeginTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
		fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
		deployer.On("Deploy", fn).Return(nil)

		computes := 0
		three := int32(3)
		topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &three}}
		deployer.On("Deploy", fn).Return(nil)

		tracker.On("Compute").Return(lag(fn, 2, 6, 0)).Run(func(args mock.Arguments) {
			computes++
		}).Times(5)
		tracker.On("Compute").Return(lag(fn, 2, 6, 0)).Once().Run(func(args mock.Arguments) {
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
		tracker.On("Compute").Return(lag(fn, 2, 6, 0)).Run(func(args mock.Arguments) {
			computes++
		}).Once()
		deployer.On("Scale", fn, 2).Return(nil).Once()
		deployer.On("Scale", fn, 2).Return(nil).Run(func(args mock.Arguments) {
			Expect(computes).To(Equal(7))
			closeCh <- struct{}{}
		})
		functionHandlers.AddFunc(fn)
		topicHandlers.AddFunc(topic)

		ctrl.Run(closeCh)
	})
})

func lag(fn *v1.Function, lag ...int) map[controller.Subscription]controller.PartitionedOffsets {
	result := make(map[controller.Subscription]controller.PartitionedOffsets)
	offsets := make(controller.PartitionedOffsets, len(lag))
	for i, l := range lag {
		offsets[int32(i)] = controller.Offsets{End: int64(l)}
	}
	result[controller.Subscription{Group: fn.Name, Topic: fn.Spec.Input}] = offsets
	return result
}
