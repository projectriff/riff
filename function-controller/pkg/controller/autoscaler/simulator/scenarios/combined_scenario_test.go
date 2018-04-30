package scenarios

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
			rm.DesireReplicas(9)
			for i := 0; i < replicaInitialisationDelaySteps-1; i++ {
				rm.Tick()
				Expect(rm.actual).To(Equal(0))
			}
			rm.Tick()
			Expect(rm.actual).To(Equal(9))
		})

		It("should delay new increases", func() {
			rm.DesireReplicas(9)
			for i := 0; i < replicaInitialisationDelaySteps-1; i++ {
				rm.Tick()
				Expect(rm.actual).To(Equal(0))
				rm.DesireReplicas(12)
			}
			rm.Tick()
			Expect(rm.actual).To(Equal(9))
			rm.Tick()
			Expect(rm.actual).To(Equal(12))
		})

		It("should expedite decreases", func() {
			rm.DesireReplicas(9)
			rm.DesireReplicas(5)

			for i := 0; i < replicaInitialisationDelaySteps-1; i++ {
				rm.Tick()
				Expect(rm.actual).To(Equal(0))
			}
			rm.Tick()
			Expect(rm.actual).To(Equal(5))
		})

	})

})
