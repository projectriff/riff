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
	"testing"
	"time"

	"github.com/projectriff/function-controller/mocks"
	"github.com/projectriff/function-controller/pkg/controller"
	"github.com/projectriff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type ControllerTestSuite struct {
	suite.Suite
	ctrl                controller.Controller
	tracker             *mocks.LagTracker
	deployer            *mocks.Deployer
	deploymentsHandlers cache.ResourceEventHandlerFuncs
	functionHandlers    cache.ResourceEventHandlerFuncs
	topicHandlers       cache.ResourceEventHandlerFuncs
	closeCh             chan struct{}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (suite *ControllerTestSuite) SetupTest() {
	suite.tracker = new(mocks.LagTracker)
	suite.deployer = new(mocks.Deployer)

	topicInformer := new(mocks.TopicInformer)
	functionInformer := new(mocks.FunctionInformer)
	deploymentInformer := new(mocks.DeploymentInformer)

	siiTopics := new(mocks.SharedIndexInformer)
	topicInformer.On("Informer").Return(siiTopics)
	siiTopics.On("AddEventHandler", mock.AnythingOfType("cache.ResourceEventHandlerFuncs")).Run(func(args mock.Arguments) {
		suite.topicHandlers = args.Get(0).(cache.ResourceEventHandlerFuncs)
	})
	siiTopics.On("Run", mock.Anything)

	siiFunctions := new(mocks.SharedIndexInformer)
	functionInformer.On("Informer").Return(siiFunctions)
	siiFunctions.On("AddEventHandler", mock.AnythingOfType("cache.ResourceEventHandlerFuncs")).Run(func(args mock.Arguments) {
		suite.functionHandlers = args.Get(0).(cache.ResourceEventHandlerFuncs)
	})
	siiFunctions.On("Run", mock.Anything)

	siiDeployments := new(mocks.SharedIndexInformer)
	deploymentInformer.On("Informer").Return(siiDeployments)
	siiDeployments.On("AddEventHandler", mock.AnythingOfType("cache.ResourceEventHandlerFuncs")).Run(func(args mock.Arguments) {
		suite.deploymentsHandlers = args.Get(0).(cache.ResourceEventHandlerFuncs)
	})
	siiDeployments.On("Run", mock.Anything)

	suite.ctrl = controller.New(topicInformer, functionInformer, deploymentInformer, suite.deployer, suite.tracker)
	suite.closeCh = make(chan struct{}, 2) // 2 allows to easily send in a .Runt() func() on stubs w/o blocking
}

func (suite *ControllerTestSuite) TearDownTest() {
	suite.tracker.AssertExpectations(suite.T())
	suite.deployer.AssertExpectations(suite.T())
}

// =====================================================================================================================

func (suite *ControllerTestSuite) TestEmptyControllerAndProperShutdown() {
	suite.tracker.On("Compute").Return(nil).Run(func(args mock.Arguments) {
		suite.closeCh <- struct{}{}
	})
	suite.ctrl.Run(suite.closeCh)
}

func (suite *ControllerTestSuite) TestFunctionsComingAndGoing() {
	suite.ctrl.SetScalingInterval(10 * time.Millisecond)

	suite.tracker.On("BeginTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
	suite.tracker.On("StopTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
	fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
	suite.deployer.On("Deploy", fn).Return(nil)
	suite.tracker.On("Compute").Return(lag(fn, 1))
	suite.deployer.On("Scale", fn, 1).Return(nil).Run(func(args mock.Arguments) {
		suite.functionHandlers.DeleteFunc(fn)
	})
	suite.deployer.On("Undeploy", fn).Return(nil).Run(func(args mock.Arguments) {
		suite.closeCh <- struct{}{}
	})
	suite.functionHandlers.AddFunc(fn)

	suite.ctrl.Run(suite.closeCh)
}

func (suite *ControllerTestSuite) TestFunctionWithNonTrivialInputTopic() {
	suite.ctrl.SetScalingInterval(10 * time.Millisecond)

	suite.tracker.On("BeginTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
	fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
	suite.deployer.On("Deploy", fn).Return(nil)

	three := int32(3)
	topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &three}}
	suite.deployer.On("Deploy", fn).Return(nil)

	suite.tracker.On("Compute").Return(lag(fn, 1, 0, 0)).Once()
	suite.tracker.On("Compute").Return(lag(fn, 6, 0, 1)).Once()
	suite.tracker.On("Compute").Return(lag(fn, 2, 3, 10)).Once()
	suite.deployer.On("Scale", fn, 1).Return(nil)
	suite.deployer.On("Scale", fn, 2).Return(nil)
	suite.deployer.On("Scale", fn, 3).Return(nil).Run(func(args mock.Arguments) {
		suite.closeCh <- struct{}{}
	})
	suite.functionHandlers.AddFunc(fn)
	suite.topicHandlers.AddFunc(topic)

	suite.ctrl.Run(suite.closeCh)
}

// Tests that when actual replicas are disrupted (for whatever reason), the controller eventually requests a scaling
// to the number of replicas it thinks is correct
func (suite *ControllerTestSuite) TestReplicasReconciliation() {
	suite.ctrl.SetScalingInterval(10 * time.Millisecond)

	suite.tracker.On("BeginTracking", controller.Subscription{Topic: "input", Group: "fn"}).Return(nil)
	fn := &v1.Function{ObjectMeta: metav1.ObjectMeta{Name: "fn"}, Spec: v1.FunctionSpec{Input: "input"}}
	suite.deployer.On("Deploy", fn).Return(nil)

	computes := 0
	three := int32(3)
	topic := &v1.Topic{ObjectMeta: metav1.ObjectMeta{Name: "input"}, Spec: v1.TopicSpec{Partitions: &three}}
	suite.deployer.On("Deploy", fn).Return(nil)

	suite.tracker.On("Compute").Return(lag(fn, 2, 6, 0)).Run(func(args mock.Arguments) {
		computes++
	}).Times(5)
	suite.tracker.On("Compute").Return(lag(fn, 2, 6, 0)).Once().Run(func(args mock.Arguments) {
		computes++
		// Disrupt actual replicas on 6th computation
		deployment := v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "fn",
				Labels: map[string]string{"function": "fn"},
			},
			Status: v1beta1.DeploymentStatus{Replicas: int32(6)},
		}
		suite.deploymentsHandlers.UpdateFunc(&deployment, &deployment)
	})
	suite.tracker.On("Compute").Return(lag(fn, 2, 6, 0)).Run(func(args mock.Arguments) {
		computes++
	}).Once()
	suite.deployer.On("Scale", fn, 2).Return(nil).Once()
	suite.deployer.On("Scale", fn, 2).Return(nil).Run(func(args mock.Arguments) {
		suite.Require().Equal(7, computes)
		suite.closeCh <- struct{}{}
	})
	suite.functionHandlers.AddFunc(fn)
	suite.topicHandlers.AddFunc(topic)

	suite.ctrl.Run(suite.closeCh)
}

func lag(fn *v1.Function, lag ...int) map[controller.Subscription]controller.PartitionedOffsets {
	result := make(map[controller.Subscription]controller.PartitionedOffsets)
	offsets := make(controller.PartitionedOffsets, len(lag))
	for i, l := range lag {
		offsets[int32(i)] = controller.Offsets{End: int64(l)}
	}
	result[controller.Subscription{Group: fn.Name, Topic: fn.Spec.Input}] = offsets
	return result
}
