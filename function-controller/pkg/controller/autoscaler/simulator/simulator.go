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

const (
	simulationSteps = 1000
	replicaInitialisationDelaySteps = 15 // delayed initialisation (termination is immediate)
)

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

	actualReplicas := 0

	rm := &replicaModel{}

	for i := 0; i < simulationSteps; i++ {
		simUpdater.UpdateProducerFor(i, &queueLen)
		simUpdater.UpdatedConsumerFor(i, actualReplicas, &queueLen)

		scalerOutput := scaler.Propose()

		scalerOutputForStubFunction := scalerOutput[stubFunctionID]

		rm.desireReplicas(scalerOutputForStubFunction)
		rm.tick()
		actualReplicas = rm.actualReplicas()
		// Scale number of replicas
		step := fmt.Sprintf("%d %d %d\n", i, actualReplicas, queueLen)
		dataFile.WriteString(step)

		scaler.InformFunctionReplicas(stubFunctionID, actualReplicas)
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

// A replicaModel models the way new replicas take a while to start.
// An increase of N in the desired number of replicas results in N items being added to `scheduled`. Each item in
// `scheduled` represents a time (in "ticks") at which the corresponding replica will count towards the actual number of
// replicas. This isn't quite the same behaviour as k8s, which knows about, and will inform the function controller of,
// replicas which are still completing their initialisation, but at least it makes the model more realistic than if
// replicas initialise instantaneously.
// A decrease in the desired number of replicas is acted upon immediately by removing items from `scheduled` and, if
// that isn't sufficient, reducing the actual number of replicas.
type replicaModel struct {
	currentTime int // in "ticks"
	actual int
	lastDesired int
	scheduled []int
}

func (rm *replicaModel) desireReplicas(desired int) {
	if desired == rm.lastDesired {
		return
	}
	if desired > rm.lastDesired {
		// schedule some new replicas with a delay
		initTime := rm.currentTime + replicaInitialisationDelaySteps
		for i := desired-rm.lastDesired; i > 0; i-- {
			rm.scheduled = append(rm.scheduled, initTime)
		}
	} else {
		rm.trim(rm.lastDesired - desired)
	}
	rm.lastDesired = desired
}

func (rm *replicaModel) trim(deschedule int) {
	if deschedule >= len(rm.scheduled) {
		trimmed := len(rm.scheduled)
		rm.scheduled = []int{}	
		rm.actual -= deschedule - trimmed
	} else {
		rm.scheduled = rm.scheduled[0:len(rm.scheduled) - deschedule]
	}
}

func (rm *replicaModel) actualReplicas() int {
	return rm.actual
}

func (rm *replicaModel) tick() {
	rm.currentTime++
	remaining := []int{}
	for _, s := range rm.scheduled {
		if s <= rm.currentTime {
			rm.actual++
		} else {
			remaining = append(remaining, s)
		}
	}
	rm.scheduled = remaining
}
