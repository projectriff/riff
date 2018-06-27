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
	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler/simulator/scenarios"
)

const (
	simulationSteps = 10000
	maxReplicas     = 1000000 // avoid setting this too high to avoid int overflow during scaling calculation. 1000000 is a reasonable high value.
)

func main() {
	fmt.Println("starting to drive autoscaler with simulated workload")

	dataFile, err := os.Create("scaler.dat")
	if err != nil {
		panic(err)
	}
	defer dataFile.Close()

	stubFunctionID := autoscaler.LinkId{Link: "stub function"}

	receiver, simUpdater, rm := scenarios.MakeNewCombinedScenario()
	queueLen := int64(0)
	inspector := newStubInspector(&queueLen)

	scaler := autoscaler.NewAutoScaler(receiver, inspector)
	defer scaler.Close()

	scaler.SetMaxReplicasPolicy(func(function autoscaler.LinkId) int {
		return maxReplicas
	})

	scaler.Run()
	scaler.StartMonitoring("topic", stubFunctionID)

	actualReplicas := 0

	writes := 0

	for i := 0; i < simUpdater.SimulationSteps(); i++ {
		simUpdater.UpdateProducerFor(receiver, i, &queueLen, &writes)
		simUpdater.UpdatedConsumerFor(receiver, i, actualReplicas, &queueLen)

		scalerOutput := scaler.Propose()

		scalerOutputForStubFunction := scalerOutput[stubFunctionID]

		rm.DesireReplicas(scalerOutputForStubFunction)
		rm.Tick()
		actualReplicas = rm.ActualReplicas()
		// Scale number of replicas
		step := fmt.Sprintf("%d %d %d %d\n", i, actualReplicas, queueLen, writes)
		dataFile.WriteString(step)

		scaler.InformFunctionReplicas(stubFunctionID, actualReplicas)
	}

	fmt.Println("simulation completed")
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
