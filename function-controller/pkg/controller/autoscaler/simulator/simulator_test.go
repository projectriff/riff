package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Simulator", func() {
	Describe("replicaModel", func() {
		var (
			rm *replicaModel
		)

		BeforeEach(func() {
			rm = &replicaModel{}
		})

		It("should delay increasing the actual number of replicas", func() {
			rm.desireReplicas(9)
			for i := 0; i < replicaInitialisationDelaySteps-1; i++ {
				rm.tick()
				Expect(rm.actual).To(Equal(0))
			}
			rm.tick()
			Expect(rm.actual).To(Equal(9))
		})

		It("should delay new increases", func() {
			rm.desireReplicas(9)
			for i := 0; i < replicaInitialisationDelaySteps-1; i++ {
				rm.tick()
				Expect(rm.actual).To(Equal(0))
				rm.desireReplicas(12)
			}
			rm.tick()
			Expect(rm.actual).To(Equal(9))
			rm.tick()
			Expect(rm.actual).To(Equal(12))
		})

		It("should expedite decreases", func() {
			rm.desireReplicas(9)
			rm.desireReplicas(5)

			for i := 0; i < replicaInitialisationDelaySteps-1; i++ {
				rm.tick()
				Expect(rm.actual).To(Equal(0))
			}
			rm.tick()
			Expect(rm.actual).To(Equal(5))
		})

	})

})
