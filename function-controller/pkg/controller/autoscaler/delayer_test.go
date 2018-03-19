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

package autoscaler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/function-controller/pkg/controller/autoscaler"
	"time"
)

var _ = Describe("Delayer", func() {

	var (
		delayer *autoscaler.Delayer
	)

	BeforeEach(func() {
		delayer = autoscaler.NewDelayer(func() time.Duration {
			return time.Millisecond * 100
		})
	})

	It("should initially propose zero", func() {
		Expect(delayer.Get()).To(Equal(0))
	})

	Context("when scaled up", func() {
		BeforeEach(func() {
			delayer.Delay(1)
		})

		It("should immediately return the scaled up value", func() {
			Expect(delayer.Get()).To(Equal(1))
		})

		Context("when scaled down to 0", func() {
			BeforeEach(func() {
				delayer.Delay(0)
			})

			It("should continue to return the scaled up value", func() {
				Consistently(func() int { return delayer.Get(); }, time.Millisecond*50).Should(Equal(1))
			})

			It("should eventually return the scaled down value", func() {
				Eventually(func() int { return delayer.Get(); }, time.Millisecond*150).Should(Equal(0))
			})
		})
	})

	Context("when scaled up", func() {
		BeforeEach(func() {
			delayer.Delay(10)
		})

		It("should immediately return the scaled up value", func() {
			Expect(delayer.Get()).To(Equal(10))

		})

		Context("when scaled down in stages but not to 0", func() {
			BeforeEach(func() {
				delayer.Delay(8)
				delayer.Delay(4)
			})

			It("should immediately return the scaled down value", func() {
				Expect(delayer.Get()).To(Equal(4))
			})
		})

		Context("when scaled down to 0 but with a 'blip'", func() {
			BeforeEach(func() {
				delayer.Delay(0)
				time.Sleep(time.Millisecond * 10)
				delayer.Delay(9)
				time.Sleep(time.Millisecond * 10)
				delayer.Delay(0)
			})

			It("should continue to return the blip value", func() {
				Consistently(func() int { return delayer.Get(); }, time.Millisecond*50).Should(Equal(9))
			})
		})
	})
})
