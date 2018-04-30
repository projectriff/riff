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
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackoff(t *testing.T) {
	as := assert.New(t)
	b := &Backoff{
		Duration:   1,
		Multiplier: 2,
		MaxRetries: 3,
	}
	for i := 0; i < b.MaxRetries; i++ {
		as.True(b.Backoff())
		as.Equal(i+1, b.retries)
		as.Equal(time.Duration(math.Pow(float64(b.Multiplier),
			float64(i)))*time.Millisecond, b.duration)
	}
	as.False(b.Backoff())
}

func TestBackoffValidation(t *testing.T) {
	as := assert.New(t)
	b := new(Backoff)
	err := b.IsValid()
	as.Error(err)

	b = &Backoff{
		Duration:   1,
		Multiplier: 0,
		MaxRetries: 3,
	}
	err = b.IsValid()
	as.Error(err)

}
