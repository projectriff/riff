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
	Describe("A valid Backoff configuration", func() {
		var b *Backoff
		BeforeEach(func() {
			b, _ = NewBackoff(1*time.Millisecond, 3, 2)
		})
		Context("With maxRetries set", func() {
			It("should backoff maxRetries times", func() {
				for i := 0; i < b.maxRetries; i++ {
					Expect(b.Backoff()).To(BeTrue())
				}
				Expect(b.Backoff()).To(BeFalse())
			})
		})
		Context("With multiplier set", func() {
			It("should exponentially increase duration", func() {
				for i := 0; i < b.maxRetries; i++ {
					b.Backoff()
					Expect(b.duration).To(Equal(time.Duration(math.Pow(float64(b.multiplier),
						float64(i))) * time.Millisecond))
				}
			})
		})
		Describe("a NewBackoff", func() {
			Context("With duration = 0 ", func() {
				It("should return a duration error", func() {
					b, err := NewBackoff(0*time.Millisecond, 3, 2)
					Expect(b).To(BeNil())
					Expect(err.Error()).To(Equal("'duration' must be > 0"))
				})
			})
			Context("With duration < 0 ", func() {
				It("should return a duration error", func() {
					b, err := NewBackoff(-1*time.Millisecond, 3, 2)
					Expect(b).To(BeNil())
					Expect(err.Error()).To(Equal("'duration' must be > 0"))
				})
			})
			Context("With maxRetries = 0 ", func() {
				It("should return a maxRetries error", func() {
					b, err := NewBackoff(1*time.Millisecond, 0, 2)
					Expect(b).To(BeNil())
					Expect(err.Error()).To(Equal("'maxRetries' must be > 0"))
				})
			})
			Context("With maxRetries < 0 ", func() {
				It("should return a maxRetries error", func() {
					b, err := NewBackoff(1*time.Millisecond, -1, 2)
					Expect(b).To(BeNil())
					Expect(err.Error()).To(Equal("'maxRetries' must be > 0"))
				})
			})
			Context("With multiplier = 0 ", func() {
				It("should return a multiplier error", func() {
					b, err := NewBackoff(1*time.Millisecond, 1, 0)
					Expect(b).To(BeNil())
					Expect(err.Error()).To(Equal("'multiplier' must be > 0"))
				})
			})
			Context("With multiplier < 0 ", func() {
				It("should return a multiplier error", func() {
					b, err := NewBackoff(1*time.Millisecond, 1, -1)
					Expect(b).To(BeNil())
					Expect(err.Error()).To(Equal("'multiplier' must be > 0"))
				})
			})

		})
	})
})
