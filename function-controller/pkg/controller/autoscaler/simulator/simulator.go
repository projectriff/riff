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

package simulator

import (
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

type SimulationUpdater interface {
	UpdateProducerFor(receiver metrics.MetricsReceiver, simulationRound int, queueLen *int64, writes *int)
	UpdatedConsumerFor(receiver metrics.MetricsReceiver, simulationRound int, replicas int, queueLen *int64)
}

type ReplicaModel interface {
	DesireReplicas(int)
	ActualReplicas() int
	Tick()
}

type SimulationConfiguration interface {
	MakeNewSimulation() (metrics.MetricsReceiver, SimulationUpdater, ReplicaModel)
}
