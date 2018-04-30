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
package backoff

import (
	"errors"
	"strings"
	"time"
)

type Backoff struct {
	MaxRetries, retries, Multiplier, Duration int
	duration                                  time.Duration
}

func (b Backoff) IsValid() error {
	msgs := []string{}

	if b.MaxRetries <= 0 {
		msgs = append(msgs, "'MaxRetries' must be > 0")
	}
	if b.Multiplier <= 0 {
		msgs = append(msgs, "'Multiplier' must be > 0")
	}
	if b.Duration <= 0 {
		msgs = append(msgs, "'Duration' must be > 0")
	}
	if len(msgs) == 0 {
		return nil
	}
	return errors.New(strings.Join(msgs, ", "))

}

func (b *Backoff) Backoff() bool {
	// Back off a bit to give the invoker time to come back
	// (if we support windowing or polling this logic will be more complex)

	if b.retries > 0 {
		b.duration = b.duration * time.Duration(b.Multiplier)
	} else {
		b.duration = time.Duration(b.Duration) * time.Millisecond
	}

	b.retries++
	if b.retries <= b.MaxRetries {
		time.Sleep(b.duration)
		return true
	}
	return false
}
