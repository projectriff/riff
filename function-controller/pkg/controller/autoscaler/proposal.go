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

type Proposal struct {
	delayPolicy   func() time.Duration
	value         int
	previousValue int
	delayDeadline time.Time
}

// NewProposal creates a proposal for which scaling down only takes effect after the delay returned by the given policy.
func NewProposal(delayPolicy func() time.Duration) *Proposal {
	return &Proposal{
		delayPolicy: delayPolicy,
	}
}

func (b *Proposal) Propose(proposal int) {
	if proposal < 0 {
		panic("Proposed numbers of replicas must be non-negative")
	}

	now := time.Now()
	if proposal > 0 {
		b.value = proposal
		b.previousValue = proposal
		b.delayDeadline = now
	} else if proposal < b.value {
		// Defer scaling to 0
		if now.After(b.delayDeadline) {
			b.delayDeadline = now.Add(b.delayPolicy())
			b.previousValue = b.value
		} else {
			// Include this scaling down in the previous delay
		}
		b.value = proposal
	}
}

func (b *Proposal) Get() int {
	if time.Now().Before(b.delayDeadline) {
		return b.previousValue
	}
	return b.value
}
