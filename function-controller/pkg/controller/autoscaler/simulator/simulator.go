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

package main

import (
	"fmt"
	"os"

	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
	"time"
)

const simulationSteps = 1000

func main() {
	fmt.Println("starting to drive autoscaler with simulated workload")

	dataFile, err := os.Create("scaler.dat")
	if err != nil {
		panic(err)
	}
	defer dataFile.Close()

	stubFunctionID := autoscaler.FunctionId{Function: "stub function"}

	receiver := newStubReceiver()
	queueLen := int64(0)
	inspector := newStubInspector(&queueLen)

	scaler := autoscaler.NewAutoScaler(receiver, inspector)
	defer scaler.Close()

	scaler.SetMaxReplicasPolicy(func(function autoscaler.FunctionId) int {
		return 1000000 // avoid int overflow during scaling calculation
	})

	var simUpdater SimulationUpdater
	simUpdater = receiver

	scaler.Run()
	scaler.StartMonitoring("topic", stubFunctionID)

	replicas := 0


	for i := 0; i < simulationSteps; i++ {
		simUpdater.UpdateProducerFor(i, &queueLen)
		simUpdater.UpdatedConsumerFor(i, replicas, &queueLen)

		scalerOutput := scaler.Propose()

		scalerOutputForStubFunction := scalerOutput[stubFunctionID]
		//step := fmt.Sprintf("%d %d\n", i, scalerOutputForStubFunction)
		step := fmt.Sprintf("%d %d %d\n", i, scalerOutputForStubFunction, queueLen)
		dataFile.WriteString(step)

		// Scale number of replicas with no delay
		replicas = scalerOutputForStubFunction
		scaler.InformFunctionReplicas(stubFunctionID, scalerOutputForStubFunction)
	}

	fmt.Println("simulation completed")
}

type SimulationUpdater interface {
	UpdateProducerFor(simulationRound int, queueLen *int64)
	UpdatedConsumerFor(simulationRound int, replicas int, queueLen *int64)
}

type stubReceiver struct {
	producerMetricsChan chan metrics.ProducerAggregateMetric
	consumerMetricsChan chan metrics.ConsumerAggregateMetric

	currentRound int
}

func (rec *stubReceiver) ProducerMetrics() <-chan metrics.ProducerAggregateMetric {
	return rec.producerMetricsChan
}

func (rec *stubReceiver) ConsumerMetrics() <-chan metrics.ConsumerAggregateMetric {
	return rec.consumerMetricsChan
}

func (rec *stubReceiver) UpdateProducerFor(simulationRound int, queueLen *int64) {
	numToWrite := 0
	if simulationRound < 100 {
		numToWrite = 20
	} else if simulationRound < 500 {
		numToWrite = 40
	} else if simulationRound < 750 {
		numToWrite = 0
	} else {
		numToWrite = 20
	}
	for i := numToWrite; i > 0; i-- {
		rec.producerMetricsChan <- metrics.ProducerAggregateMetric{Topic: "topic", Count: 1}
		(*queueLen)++
	}
}

func (rec *stubReceiver) UpdatedConsumerFor(simulationRound int, replicas int, queueLen *int64) {
	if replicas == 0 {
		return // nothing doing
	}

	numToRead := 10 * replicas
	if *queueLen < int64(numToRead) {
		numToRead = int(*queueLen)
	}
	for i := numToRead; i > 0; i-- {
		rec.consumerMetricsChan <- metrics.ConsumerAggregateMetric{Topic: "topic", ConsumerGroup: "stub function", Pod: "pod" /*should vary by replica*/ , Count: 1, Interval: time.Millisecond}
		(*queueLen)--
	}
}

func newStubReceiver() *stubReceiver {
	return &stubReceiver{
		producerMetricsChan: make(chan metrics.ProducerAggregateMetric),
		consumerMetricsChan: make(chan metrics.ConsumerAggregateMetric),
	}
}

type stubInspector struct {
	queueLen *int64
}

func newStubInspector(queueLen *int64) *stubInspector {
	return &stubInspector{
		queueLen: queueLen,
	}
}

func (i *stubInspector) QueueLength(topic string, function string) (int64, error) {
	return *i.queueLen, nil
}
