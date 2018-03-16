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

var _ = Describe("Proposal", func() {

	var (
		proposal *autoscaler.Proposal
	)

	BeforeEach(func() {
		proposal = autoscaler.NewProposal(func() time.Duration {
			return time.Millisecond * 100
		})
	})

	It("should initially propose zero", func() {
		Expect(proposal.Get()).To(Equal(0))
	})

	Context("when scaled up", func() {
		BeforeEach(func() {
			proposal.Propose(1)
		})

		It("should immediately propose the scaled up value", func() {
			Expect(proposal.Get()).To(Equal(1))
		})

		Context("when scaled down to 0", func() {
			BeforeEach(func() {
				proposal.Propose(0)
			})

			It("should continue to propose the scaled up value", func() {
				Consistently(func() int { return proposal.Get(); }, time.Millisecond*50).Should(Equal(1))
			})

			It("should eventually propose the scaled down value", func() {
				Eventually(func() int { return proposal.Get(); }, time.Millisecond*150).Should(Equal(0))
			})
		})
	})

	Context("when scaled up", func() {
		BeforeEach(func() {
			proposal.Propose(10)
		})

		It("should immediately propose the scaled up value", func() {
			Expect(proposal.Get()).To(Equal(10))

		})

		Context("when scaled down in stages but not to 0", func() {
			BeforeEach(func() {
				proposal.Propose(8)
				proposal.Propose(4)
			})

			It("should immediately propose the scaled down value", func() {
				Expect(proposal.Get()).To(Equal(4))
			})
		})

		Context("when scaled down to 0 but with a 'blip'", func() {
			BeforeEach(func() {
				proposal.Propose(0)
				time.Sleep(time.Millisecond * 10)
				proposal.Propose(9)
				time.Sleep(time.Millisecond * 10)
				proposal.Propose(0)
			})

			It("should continue to propose the blip value", func() {
				Consistently(func() int { return proposal.Get(); }, time.Millisecond*50).Should(Equal(9))
			})
		})
	})
})
