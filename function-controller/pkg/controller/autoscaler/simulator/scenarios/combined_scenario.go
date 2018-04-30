package scenarios

import (
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler/simulator"
	"math"
	"time"
)

const (
	replicaInitialisationDelaySteps = 15 // 1.5 s delayed initialisation (termination is immediate)
	containerPullDelaySteps         = 0  // 50 // 5.0 s extra delay in starting the first replica

	maxWritesPerTick = 40
)

type CombinedScenario struct {}

func (scenario CombinedScenario) MakeNewSimulation() (metrics.MetricsReceiver, simulator.SimulationUpdater, simulator.ReplicaModel) {
	stubReceiver := newStubReceiver()
	rm := &replicaModel{initialDelay: containerPullDelaySteps}

	return stubReceiver, stubReceiver, rm
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

func (rec *stubReceiver) UpdateProducerFor(simulationRound int, queueLen *int64, writes *int) {
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


// A replicaModel models the way new replicas take a while to start.
// An increase of N in the desired number of replicas results in N items being added to `scheduled`. Each item in
// `scheduled` represents a time (in "ticks") at which the corresponding replica will count towards the actual number of
// replicas. This isn't quite the same behaviour as k8s, which knows about, and will inform the function controller of,
// replicas which are still completing their initialisation, but at least it makes the model more realistic than if
// replicas initialise instantaneously.
// A decrease in the desired number of replicas is acted upon immediately by removing items from `scheduled` and, if
// that isn't sufficient, reducing the actual number of replicas.
type replicaModel struct {
	currentTime  int // in "ticks"
	actual       int
	lastDesired  int
	scheduled    []int
	initialDelay int
}

func (rm *replicaModel) DesireReplicas(desired int) {
	if desired == rm.lastDesired {
		return
	}
	if desired > rm.lastDesired {
		// schedule some new replicas with a delay
		initTime := rm.currentTime + replicaInitialisationDelaySteps + rm.initialDelay
		rm.initialDelay = 0
		for i := desired - rm.lastDesired; i > 0; i-- {
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
		rm.scheduled = rm.scheduled[0 : len(rm.scheduled)-deschedule]
	}
}

func (rm *replicaModel) ActualReplicas() int {
	return rm.actual
}

func (rm *replicaModel) Tick() {
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
