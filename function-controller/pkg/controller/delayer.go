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

import (
	"log"
	"time"
)

type delayer struct {
	*ctrl
	scaleDownToZeroDecision   map[fnKey]time.Time
	previousCombinedPositions activityCounts
}

// delay transforms passed in replica counts to add some delay (this is achieved by returning the current number of
// replicas rather than the 'in' value) in the following case:
// - when scaling to 0, let some idleTimeout pass. The clock is reset if there is activity
// - when scaling UP, don't scale up further until at least some activity has been recorded
func (c *delayer) delay(in replicaCounts, combinedPositions activityCounts) (replicaCounts, activityCounts) {
	result := make(map[fnKey]int32, len(in))
	now := time.Now()
	for fn, target := range in {
		result[fn] = in[fn]
		if target == 0 && c.actualReplicas[fn] > 0 {
			if start, ok := c.scaleDownToZeroDecision[fn]; ok {
				//idleTimeout := 10 * time.Second
				idleTimeout := time.Duration(*c.functions[fn].Spec.IdleTimeoutMs) * time.Millisecond
				if now.Before(start.Add(idleTimeout)) { // Timeout not elapsed, don't proceed with scale down
					result[fn] = c.actualReplicas[fn]
					log.Printf("Still waiting %v for %v", start.Add(idleTimeout).Sub(now), fn.name)
					if combinedPositions[fn].end > c.previousCombinedPositions[fn].end { // Still some activity, reset the clock
						log.Printf("Resetting the clock for %v: %v, %v", fn.name, combinedPositions[fn], c.previousCombinedPositions[fn])
						delete(c.scaleDownToZeroDecision, fn)
					}
				} else {
					delete(c.scaleDownToZeroDecision, fn)
				}
			} else {
				c.scaleDownToZeroDecision[fn] = now
				result[fn] = c.actualReplicas[fn]
			}
		} else if target > c.actualReplicas[fn] && c.actualReplicas[fn] > 0 { // Scaling UP
			delete(c.scaleDownToZeroDecision, fn)
			if pcc, ok := c.previousCombinedPositions[fn]; ok && pcc == combinedPositions[fn] {
				// No activity since last tick. Must be rebalancing. Wait before scaling even more
				result[fn] = c.actualReplicas[fn]
			}

		} else if target < c.actualReplicas[fn] { // Scaling DOWN in the 1-N range
			delete(c.scaleDownToZeroDecision, fn)
		}
	}
	c.previousCombinedPositions = combinedPositions
	return result, combinedPositions
}
