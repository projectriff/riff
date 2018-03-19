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

package autoscaler

import "time"

type Delayer struct {
	policy        func() time.Duration
	value         int
	previousValue int
	deadline      time.Time
}

// NewDelayer creates a delayer which delays scaling down to zero until the delay returned by the given policy has elapsed.
func NewDelayer(delayPolicy func() time.Duration) *Delayer {
	return &Delayer{
		policy: delayPolicy,
	}
}

func (b *Delayer) Delay(proposal int) *Delayer {
	if proposal < 0 {
		panic("Proposed numbers of replicas must be non-negative")
	}

	now := time.Now()
	if proposal > 0 {
		b.value = proposal
		b.previousValue = proposal
		b.deadline = now
	} else if b.value > 0 {
		// Defer scaling to 0
		if now.After(b.deadline) {
			b.deadline = now.Add(b.policy())
			b.previousValue = b.value
		} else {
			// Include this scaling down in the previous delay
		}
		b.value = proposal
	}

	return b
}

func (b *Delayer) Get() int {
	if time.Now().Before(b.deadline) {
		return b.previousValue
	}
	return b.value
}
