package scenarios

import (
	"math"
	"time"

	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler/simulator"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

type combinedScenario struct{}

func MakeNewCombinedScenario() (metrics.MetricsReceiver, simulator.SimulationUpdater, simulator.ReplicaModel) {
	stubReceiver := newStubReceiver()
	rm := &replicaModel{initialDelay: containerPullDelaySteps}
	scenario := &combinedScenario{}

	return stubReceiver, scenario, rm
}

func (scenario *combinedScenario) UpdateProducerFor(receiver metrics.MetricsReceiver, simulationRound int, queueLen *int64, writes *int) {
	stubReceiver := receiver.(stubReceiver)

	numToWrite := 0
	if simulationRound < 100 {
		// initial quiet interval
		numToWrite = 0
	} else if simulationRound < 1000 {
		// step up
		numToWrite = maxWritesPerTick / 2
	} else if simulationRound < 2000 {
		// further step up
		numToWrite = maxWritesPerTick
	} else if simulationRound < 3000 {
		// step down
		numToWrite = maxWritesPerTick / 2
	} else if simulationRound < 4000 {
		// step down to quiet interval
		numToWrite = 0
	} else if simulationRound < 6000 {
		// sinusoidal interval
		numToWrite = maxWritesPerTick/2 + int(maxWritesPerTick*math.Sin(float64((simulationRound-4000)/167))/2)
	} else if simulationRound < 7000 {
		// step down to quiet interval
		numToWrite = 0
	} else if simulationRound < 8000 {
		// ramp up
		numToWrite = maxWritesPerTick * (simulationRound - 7000) / 1000
	} else if simulationRound < 9000 {
		// ramp down
		numToWrite = maxWritesPerTick * (9000 - simulationRound) / 1000
	}

	*writes = numToWrite

	for i := numToWrite; i > 0; i-- {
		stubReceiver.producerMetricsChan <- metrics.ProducerAggregateMetric{Topic: "topic", Count: 1}
		(*queueLen)++
	}
}

func (scenario *combinedScenario) UpdatedConsumerFor(receiver metrics.MetricsReceiver, simulationRound int, replicas int, queueLen *int64) {
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

func (scenario *combinedScenario) SimulationSteps() int {
	return 10000
}
