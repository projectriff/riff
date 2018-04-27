package backoff

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
import (
	"math"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backoff", func() {
	Context("when the constructor parameters are valid", func() {
		const (
			maxRetries = 3
			multiplier = 2
		)

		var (
			initialDuration time.Duration
			b               *Backoff
			err             error
		)

		BeforeEach(func() {
			initialDuration = time.Millisecond
			b, err = NewBackoff(initialDuration, maxRetries, multiplier)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should backoff maxRetries times and then stop backing off", func() {
			for i := 0; i < maxRetries; i++ {
				Expect(b.Backoff()).To(BeTrue())
			}
			Expect(b.Backoff()).To(BeFalse())
		})

		It("should exponentially increase the backoff duration", func() {
			for i := 0; i < maxRetries; i++ {
				b.Backoff()
				Expect(b.duration).To(Equal(initialDuration * time.Duration(math.Pow(float64(multiplier), float64(i)))))
			}
		})
	})

	Describe("Constructor parameter checking", func() {
		var (
			initialDuration time.Duration
			maxRetries      int
			multiplier      int
			err             error
		)

		BeforeEach(func() {
			initialDuration = time.Millisecond
			maxRetries = 3
			multiplier = 2
		})

		JustBeforeEach(func() {
			_, err = NewBackoff(initialDuration, maxRetries, multiplier)
		})

		Context("when duration = 0 ", func() {
			BeforeEach(func() {
				initialDuration = 0 * time.Millisecond
			})

			It("should return a duration error", func() {
				Expect(err).To(MatchError("'duration' must be > 0"))
			})
		})

		Context("when duration < 0 ", func() {
			BeforeEach(func() {
				initialDuration = -1 * time.Millisecond
			})

			It("should return a duration error", func() {
				Expect(err).To(MatchError("'duration' must be > 0"))
			})
		})

		Context("when maxRetries = 0 ", func() {
			BeforeEach(func() {
				maxRetries = 0
			})

			It("should return a maxRetries error", func() {
				Expect(err).To(MatchError("'maxRetries' must be > 0"))
			})
		})

		Context("when maxRetries < 0 ", func() {
			BeforeEach(func() {
				maxRetries = -1
			})

			It("should return a maxRetries error", func() {
				Expect(err).To(MatchError("'maxRetries' must be > 0"))
			})
		})

		Context("when multiplier = 0 ", func() {
			BeforeEach(func() {
			    multiplier = 0
			})

			It("should return a multiplier error", func() {
				Expect(err).To(MatchError("'multiplier' must be > 0"))
			})
		})

		Context("when multiplier < 0 ", func() {
			BeforeEach(func() {
			    multiplier = -1
			})

			It("should return a multiplier error", func() {
				Expect(err).To(MatchError("'multiplier' must be > 0"))
			})
		})
	})
})
