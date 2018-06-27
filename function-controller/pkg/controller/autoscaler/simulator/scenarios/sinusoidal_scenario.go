package scenarios

import (
	"math"
	"time"

	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler/simulator"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

type sinusoidalScenario struct {
	simulationSteps int
}

func MakeNewSinusoidalScenario(simulationSteps int) (metrics.MetricsReceiver, simulator.SimulationUpdater, simulator.ReplicaModel) {
	stubReceiver := newStubReceiver()
	rm := &replicaModel{initialDelay: containerPullDelaySteps}
	scenario := &sinusoidalScenario{
		simulationSteps: simulationSteps,
	}

	return stubReceiver, scenario, rm
}

func (scenario *sinusoidalScenario) UpdateProducerFor(receiver metrics.MetricsReceiver, simulationRound int, queueLen *int64, writes *int) {
	stubReceiver := receiver.(stubReceiver)

	roundSegment := (simulationRound / 500)
	roundOffset := (scenario.simulationSteps / 2) * roundSegment
	adjustedRound := simulationRound - roundOffset
	rawSine := math.Sin(float64((adjustedRound / 100)))
	writeSine := rawSine * maxWritesPerTick
	clampedSine := math.Max(0, writeSine)

	numToWrite := int(clampedSine)

	*writes = numToWrite

	for i := numToWrite; i > 0; i-- {
		stubReceiver.producerMetricsChan <- metrics.ProducerAggregateMetric{Topic: "topic", Count: 1}
		(*queueLen)++
	}
}

func (scenario *sinusoidalScenario) UpdatedConsumerFor(receiver metrics.MetricsReceiver, simulationRound int, replicas int, queueLen *int64) {
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
