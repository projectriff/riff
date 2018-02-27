/*
 * Copyright 2018 the original author or authors.
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

package controller

type smoother struct {
	*ctrl
	memory map[fnKey]float32
}

// smooth adds interpolation to replica counts calculation, so that there are no sudden jumps.
func (c *smoother) smooth(in replicaCounts, combinedPositions activityCounts) (replicaCounts, activityCounts) {
	newMemory := make(map[fnKey]float32, len(in)) // recreating map is a lazy way to get rid of keys not present in 'in'
	replicas := make(map[fnKey]int32, len(in))
	for k, v := range in {
		newMemory[k] = interpolate(c.memory[k], v, 0.05)
		replicas[k] = int32(0.5 + newMemory[k])
		// Special case for the 0->1 case to have immediate scale up
		if in[k] > 0 && c.actualReplicas[k] == 0 {
			replicas[k] = 1
			newMemory[k] = 1.0
		}
	}
	c.memory = newMemory
	return replicas, combinedPositions
}

func interpolate(current float32, target int32, greed float32) float32 {
	return current + (float32(target)-current)*greed
}
