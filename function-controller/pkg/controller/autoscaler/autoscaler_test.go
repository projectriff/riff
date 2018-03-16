/*
 * Copyright 2018-Present the original author or authors.
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

package autoscaler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics/mockmetrics"
	"time"
	"fmt"
	"github.com/onsi/gomega/types"
	"github.com/projectriff/riff/message-transport/pkg/transport/mocktransport"
	"github.com/stretchr/testify/mock"
	"errors"
)

var _ = Describe("Autoscaler", func() {

	type AutoScalerInterfaces interface {
		AutoScaler

		// internal interfaces necessary to avoid test races
		receiveProducerMetric(pm metrics.ProducerAggregateMetric)
		receiveConsumerMetric(cm metrics.ConsumerAggregateMetric)
	}

	const testTopic = "test-topic"

	var (
		auto       AutoScalerInterfaces
		testFuncId FunctionId

		mockMetricsReceiver    *mockmetrics.MetricsReceiver
		mockTransportInspector *mocktransport.Inspector

		proposal map[FunctionId]int
	)

	BeforeEach(func() {
		testFuncId = FunctionId{"test-function"}

		mockMetricsReceiver = &mockmetrics.MetricsReceiver{}

		mockTransportInspector = &mocktransport.Inspector{}

		auto = NewAutoScaler(mockMetricsReceiver, mockTransportInspector)

	})

	JustBeforeEach(func() {
		proposal = auto.Propose()
	})

	Describe("autoscaling", func() {
		BeforeEach(func() {
			Expect(auto.StartMonitoring(testTopic, testFuncId)).To(Succeed())
			auto.SetDelayScaleDownPolicy(func(function string) time.Duration {
				return time.Millisecond * 20
			})
			mockTransportInspector.On("QueueLength", testTopic, mock.Anything).Return(int64(0), nil)
		})

		Context("when messages are produced but not consumed", func() {
			BeforeEach(func() {
				auto.receiveProducerMetric(metrics.ProducerAggregateMetric{
					Topic: testTopic,
					Count: 1,
				})
			})

			It("should scale up to one", func() {
				Expect(proposal).To(Propose(testFuncId, 1))
			})

			Context("when no further messages are produced for sufficiently long", func() {
				BeforeEach(func() {
					auto.Propose()
					auto.InformFunctionReplicas(testFuncId, 1)
					auto.Propose()
					time.Sleep(time.Millisecond * 30)
				})

				It("should scale down to 0", func() {
					Expect(proposal).To(Propose(testFuncId, 0))
				})
			})
		})

		Context("when messages are produced 3 times faster than they are consumed by 1 pod", func() {
			BeforeEach(func() {
				auto.InformFunctionReplicas(testFuncId, 1)

				auto.receiveProducerMetric(metrics.ProducerAggregateMetric{
					Topic: testTopic,
					Count: 3,
				})

				auto.receiveConsumerMetric(metrics.ConsumerAggregateMetric{
					Topic:    testTopic,
					Function: testFuncId.Function,
					Count:    1,
				})
			})

			It("should scale up to 3 pods", func() {
				Expect(proposal).To(Propose(testFuncId, 3))
			})

			Context("when no further messages are produced", func() {
				BeforeEach(func() {
					auto.Propose()
					auto.InformFunctionReplicas(testFuncId, 3)
					auto.Propose()
					time.Sleep(time.Millisecond * 5)
				})

				It("should not scale down prematurely", func() {
					Expect(proposal).To(Propose(testFuncId, 3))
				})

				Context("when no further messages are produced for an extended period", func() {
					BeforeEach(func() {
						auto.Propose()
						time.Sleep(time.Millisecond * 30)
					})

					It("should scale down to 0", func() {
						Expect(proposal).To(Propose(testFuncId, 0))
					})
				})
			})

			Context("when messages are then produced 10 times faster than they are consumed by 3 pods", func() {
				BeforeEach(func() {
					auto.Propose()
					auto.InformFunctionReplicas(testFuncId, 3)

					auto.receiveProducerMetric(metrics.ProducerAggregateMetric{
						Topic: testTopic,
						Count: 100,
					})

					auto.receiveConsumerMetric(metrics.ConsumerAggregateMetric{
						Topic:    testTopic,
						Function: testFuncId.Function,
						Count:    10,
					})
				})

				It("should scale up to 30 pods", func() {
					Expect(proposal).To(Propose(testFuncId, 30))
				})

				Context("when messages are then produced 10 times slower than they are consumed by 30 pods", func() {
					BeforeEach(func() {
						auto.Propose()
						auto.InformFunctionReplicas(testFuncId, 30)

						auto.receiveProducerMetric(metrics.ProducerAggregateMetric{
							Topic: testTopic,
							Count: 10,
						})

						auto.receiveConsumerMetric(metrics.ConsumerAggregateMetric{
							Topic:    testTopic,
							Function: testFuncId.Function,
							Count:    100,
						})

						auto.Propose()
						time.Sleep(time.Millisecond * 30)
					})

					It("should scale down to 3 pods", func() {
						Expect(proposal).To(Propose(testFuncId, 3))
					})
				})
			})
		})
	})

	Describe("maximumReplicasPolicy", func() {
		Context("when the maximum replicas policy returns 2 but messages are produced 3 times faster than they are consumed by 1 pod", func() {
			BeforeEach(func() {
				Expect(auto.StartMonitoring(testTopic, testFuncId)).To(Succeed())
				auto.SetDelayScaleDownPolicy(func(function string) time.Duration {
					return time.Millisecond * 20
				})

				auto.SetMaxReplicasPolicy(func(topic string, function string) int {
					return 2;
				})
				auto.InformFunctionReplicas(testFuncId, 1)

				auto.receiveProducerMetric(metrics.ProducerAggregateMetric{
					Topic: testTopic,
					Count: 3,
				})

				auto.receiveConsumerMetric(metrics.ConsumerAggregateMetric{
					Topic:    testTopic,
					Function: testFuncId.Function,
					Count:    1,
				})
			})

			It("should scale up to 2 pods (rather than 3)", func() {
				Expect(proposal).To(Propose(testFuncId, 2))
			})
		})
	})

	Describe("Queue length behaviour", func() {
		BeforeEach(func() {
			Expect(auto.StartMonitoring(testTopic, testFuncId)).To(Succeed())
			auto.SetDelayScaleDownPolicy(func(function string) time.Duration {
				return time.Millisecond * 20
			})
			auto.receiveProducerMetric(metrics.ProducerAggregateMetric{
				Topic: testTopic,
				Count: 1,
			})
			auto.Propose()
			auto.InformFunctionReplicas(testFuncId, 1)
			auto.Propose()
			time.Sleep(time.Millisecond * 30)
		})

		Context("when queue length is positive", func() {
			BeforeEach(func() {
				mockTransportInspector.On("QueueLength", testTopic, mock.Anything).Return(int64(1), nil)
			})

			It("should not scale down", func() {
				Expect(proposal).To(Propose(testFuncId, 1))
			})
		})

		Context("when queue length cannot be determined", func() {
			BeforeEach(func() {
				mockTransportInspector.On("QueueLength", testTopic, mock.Anything).Return(int64(0), errors.New("beats me guv"))
			})

			It("should not scale down, to be on the safe side", func() {
				Expect(proposal).To(Propose(testFuncId, 1))
			})
		})
	})

	Describe("monitoring lifecycle", func() {
		BeforeEach(func() {
			mockTransportInspector.On("QueueLength", testTopic, mock.Anything).Return(int64(0), nil)
		})

		It("should return an empty proposal by default", func() {
			Expect(proposal).To(BeEmpty())
		})

		Context("when monitoring a given function is started", func() {
			BeforeEach(func() {
				Expect(auto.StartMonitoring(testTopic, testFuncId)).To(Succeed())
			})

			It("should propose zero replicas by default", func() {
				Expect(proposal).To(Propose(testFuncId, 0))
			})

			It("should return a copy of the proposal map", func() {
				proposal = auto.Propose()
				replicas, ok := proposal[testFuncId]
				Expect(ok).To(BeTrue())
				Expect(replicas).To(Equal(0))

				proposal[testFuncId] = 99 // corrupt the map

				proposal = auto.Propose()
				replicas, ok = proposal[testFuncId]
				Expect(ok).To(BeTrue())
				Expect(replicas).To(Equal(0))

			})

			Context("when the function is running in one pod", func() {
				BeforeEach(func() {
					auto.InformFunctionReplicas(testFuncId, 1)
				})

				It("should propose zero replicas by default", func() {
					Expect(proposal).To(Propose(testFuncId, 0))
				})
			})

			It("should return an error when monitoring the function is started again", func() {
				Expect(auto.StartMonitoring(testTopic, testFuncId)).To(MatchError("Already monitoring topic test-topic and function {test-function}"))
			})

			Context("when monitoring of the function is stopped", func() {
				BeforeEach(func() {
					Expect(auto.StopMonitoring(testTopic, testFuncId)).To(Succeed())
				})

				It("should not propose replicas for the function by default", func() {
					Expect(proposal).NotTo(ProposeFunction(testFuncId))
				})
			})
		})

		Context("when not monitoring a given function", func() {
			It("should not propose replicas for the function by default", func() {
				Expect(proposal).NotTo(ProposeFunction(testFuncId))
			})

			It("should return an error when monitoring the function is stopped", func() {
				Expect(auto.StopMonitoring(testTopic, testFuncId)).To(MatchError("Not monitoring topic test-topic and function {test-function}"))
			})

			Context("when the function is running in one pod", func() {
				BeforeEach(func() {
					auto.InformFunctionReplicas(testFuncId, 1)
				})

				It("should not propose any replicas for the function", func() {
					Expect(proposal).NotTo(ProposeFunction(testFuncId))
				})
			})
		})

		Context("when already monitoring a given function", func() {
			var anotherFuncId FunctionId

			BeforeEach(func() {
				anotherFuncId = FunctionId{"another-function"}
				Expect(auto.StartMonitoring(testTopic, testFuncId)).To(Succeed())
			})

			Context("when monitoring the same topic and another function is started", func() {
				BeforeEach(func() {
					Expect(auto.StartMonitoring(testTopic, anotherFuncId)).To(Succeed())
				})

				It("should propose zero replicas by default", func() {
					Expect(proposal).To(Propose(anotherFuncId, 0))
				})
			})

			Context("when not monitoring another function", func() {
				It("should not propose replicas for the function by default", func() {
					Expect(proposal).NotTo(ProposeFunction(anotherFuncId))
				})
			})

			It("should return an error when monitoring another function is stopped", func() {
				Expect(auto.StopMonitoring(testTopic, anotherFuncId)).To(MatchError("Not monitoring topic test-topic and function {another-function}"))
			})
		})
	})

	// Exercise the autoscaler's goroutines minimally.
	Describe("Run", func() {
		var (
			producerMetrics chan metrics.ProducerAggregateMetric
			consumerMetrics chan metrics.ConsumerAggregateMetric
		)

		BeforeEach(func() {
			producerMetrics = make(chan metrics.ProducerAggregateMetric, 1)
			consumerMetrics = make(chan metrics.ConsumerAggregateMetric, 1)
			var pm <-chan metrics.ProducerAggregateMetric = producerMetrics
			mockMetricsReceiver.On("ProducerMetrics").Return(pm)
			var cm <-chan metrics.ConsumerAggregateMetric = consumerMetrics
			mockMetricsReceiver.On("ConsumerMetrics").Return(cm)

			auto = NewAutoScaler(mockMetricsReceiver, mockTransportInspector)
			auto.SetDelayScaleDownPolicy(func(function string) time.Duration {
				return time.Millisecond * 20
			})

			mockTransportInspector.On("QueueLength", testTopic, mock.Anything).Return(int64(0), nil)

			auto.Run()

			Expect(auto.StartMonitoring(testTopic, testFuncId)).To(Succeed())
		})

		AfterEach(func() {
			Expect(auto.Close()).To(Succeed())
		})

		Context("when metrics are produced and consumed and then not produced for a while", func() {
			BeforeEach(func() {
				producerMetrics <- metrics.ProducerAggregateMetric{
					Topic: testTopic,
				}

				consumerMetrics <- metrics.ConsumerAggregateMetric{
					Topic: testTopic,
				}

				auto.Propose()

				time.Sleep(time.Millisecond * 30)
			})

			It("should scale down to 0", func() {
				Expect(proposal).To(Propose(testFuncId, 0))
			})
		})
	})
})

func Propose(funcId FunctionId, replicas int) types.GomegaMatcher {
	return &proposeFunctionReplicasMatcher{
		funcId:   funcId,
		replicas: replicas,
	}
}

type proposeFunctionReplicasMatcher struct {
	funcId   FunctionId
	replicas int
}

func (matcher *proposeFunctionReplicasMatcher) Match(actual interface{}) (success bool, err error) {
	proposal, ok := actual.(map[FunctionId]int)
	if !ok {
		return false, fmt.Errorf("Propose matcher expects a map[FunctionId]int")
	}

	replicas, ok := proposal[matcher.funcId]
	if !ok {
		return false, nil
	}

	return replicas == matcher.replicas, nil
}

func (matcher *proposeFunctionReplicasMatcher) FailureMessage(actual interface{}) (message string) {
	proposal, ok := actual.(map[FunctionId]int)
	if !ok {
		return fmt.Sprintf("Expected proposal\n\t%#v\nto be of type map[FunctionId]int", actual)
	}

	replicas, ok := proposal[matcher.funcId]
	if !ok {
		return fmt.Sprintf("Expected proposal\n\t%#v\nto contain a number of replicas for function %#v", actual, matcher.funcId)
	}

	return fmt.Sprintf("Expected proposal for function %#v of\n\t%d replicas\nto be\n\t%d replicas", matcher.funcId, replicas, matcher.replicas)
}

func (matcher *proposeFunctionReplicasMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	proposal, ok := actual.(map[FunctionId]int)
	if !ok {
		return fmt.Sprintf("Expected proposal\n\t%#v\nto be of type map[FunctionId]int", actual)
	}

	replicas, ok := proposal[matcher.funcId]
	if !ok {
		return fmt.Sprintf("Expected proposal\n\t%#v\nnot to contain the correct number of replicas for function %#v", actual, matcher.funcId)
	}

	return fmt.Sprintf("Expected proposal for function %#v of\n\t%d replicas\nnot to be\n\t%d replicas", matcher.funcId, replicas, matcher.replicas)
}

func ProposeFunction(funcId FunctionId) types.GomegaMatcher {
	return &proposeFunctionMatcher{
		funcId: funcId,
	}
}

type proposeFunctionMatcher struct {
	funcId FunctionId
}

func (matcher *proposeFunctionMatcher) Match(actual interface{}) (success bool, err error) {
	proposal, ok := actual.(map[FunctionId]int)
	if !ok {
		return true, fmt.Errorf("ProposeFunction matcher expects a map[FunctionId]int")
	}

	_, ok = proposal[matcher.funcId]
	return ok, nil
}

func (matcher *proposeFunctionMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected proposal\n\t%#v\nto contain a number of replicas for function %#v", actual, matcher.funcId)
}

func (matcher *proposeFunctionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected proposal\n\t%#v\nnot to contain a number of replicas for function %#v", actual, matcher.funcId)
}
