package scenarios

import (
	"time"

	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler/simulator"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

type rampScenario struct {
	simulationSteps int
}

func MakeNewRampScenario(simulationSteps int) (metrics.MetricsReceiver, simulator.SimulationUpdater, simulator.ReplicaModel) {
	stubReceiver := newStubReceiver()
	rm := &replicaModel{initialDelay: containerPullDelaySteps}
	scenario := &rampScenario{
		simulationSteps: simulationSteps,
	}

	return stubReceiver, scenario, rm
}

func (scenario *rampScenario) UpdateProducerFor(receiver metrics.MetricsReceiver, simulationRound int, queueLen *int64, writes *int) {
	stubReceiver := receiver.(stubReceiver)

	numToWrite := 0
	halfway := int(float64(scenario.simulationSteps) * 0.50)
	if simulationRound < halfway {
		// ramp up
		numToWrite = maxWritesPerTick * (simulationRound) / 2000
	} else {
		// ramp down
		numToWrite = maxWritesPerTick * (scenario.simulationSteps - simulationRound) / 2000
	}

	*writes = numToWrite

	for i := numToWrite; i > 0; i-- {
		stubReceiver.producerMetricsChan <- metrics.ProducerAggregateMetric{Topic: "topic", Count: 1}
		(*queueLen)++
	}
}

func (scenario *rampScenario) UpdatedConsumerFor(receiver metrics.MetricsReceiver, simulationRound int, replicas int, queueLen *int64) {
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
