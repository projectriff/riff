package scenarios

import (
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

const (
	replicaInitialisationDelaySteps = 15 // 1.5 s delayed initialisation (termination is immediate)
	containerPullDelaySteps         = 0  // 50 // 5.0 s extra delay in starting the first replica

	maxWritesPerTick = 40
)

type stubReceiver struct {
	producerMetricsChan chan metrics.ProducerAggregateMetric
	consumerMetricsChan chan metrics.ConsumerAggregateMetric

	currentRound int
}

func (rec stubReceiver) ProducerMetrics() <-chan metrics.ProducerAggregateMetric {
	return rec.producerMetricsChan
}

func (rec stubReceiver) ConsumerMetrics() <-chan metrics.ConsumerAggregateMetric {
	return rec.consumerMetricsChan
}

func newStubReceiver() stubReceiver {
	return stubReceiver{
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
