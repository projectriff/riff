package scenarios

import (
	"time"

	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler/simulator"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

type stepScenario struct {
	simulationSteps int
}

func MakeNewStepScenario(simulationSteps int) (metrics.MetricsReceiver, simulator.SimulationUpdater, simulator.ReplicaModel) {
	stubReceiver := newStubReceiver()
	rm := &replicaModel{initialDelay: containerPullDelaySteps}
	scenario := &stepScenario{
		simulationSteps: simulationSteps,
	}

	return stubReceiver, scenario, rm
}

func (scenario *stepScenario) UpdateProducerFor(receiver metrics.MetricsReceiver, simulationRound int, queueLen *int64, writes *int) {
	stubReceiver := receiver.(stubReceiver)

	numToWrite := 0
	if simulationRound < int(float64(scenario.simulationSteps)*0.02) {
		// initial quiet interval
		numToWrite = 0
	} else if simulationRound < int(float64(scenario.simulationSteps)*0.3) {
		// step up
		numToWrite = maxWritesPerTick / 2
	} else if simulationRound < int(float64(scenario.simulationSteps)*0.6) {
		// further step up
		numToWrite = maxWritesPerTick
	} else if simulationRound < int(float64(scenario.simulationSteps)*0.98) {
		// step down
		numToWrite = maxWritesPerTick / 2
	} else if simulationRound < int(float64(scenario.simulationSteps)*0.999) {
		// step down to quiet interval
		numToWrite = 0
	}

	*writes = numToWrite

	for i := numToWrite; i > 0; i-- {
		stubReceiver.producerMetricsChan <- metrics.ProducerAggregateMetric{Topic: "topic", Count: 1}
		(*queueLen)++
	}
}

func (scenario *stepScenario) UpdatedConsumerFor(receiver metrics.MetricsReceiver, simulationRound int, replicas int, queueLen *int64) {
	stubReceiver := receiver.(stubReceiver)

	if replicas == 0 {
		return // nothing doing
	}

	numToRead := 10 * replicas
	if *queueLen < int64(numToRead) {
		numToRead = int(*queueLen)
	}
	for i := numToRead; i > 0; i-- {
		stubReceiver.consumerMetricsChan <- metrics.ConsumerAggregateMetric{Topic: "topic", ConsumerGroup: "stub function", Pod: "pod" /*should vary by replica*/, Count: 1, Interval: time.Millisecond}
		(*queueLen)--
	}
}
