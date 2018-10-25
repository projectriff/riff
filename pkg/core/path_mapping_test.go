/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/image"
)

var _ = Describe("Path Mapping", func() {
	var (
		mapper pathMapping
		name   image.Name
		mapped string
	)

	JustBeforeEach(func() {
		mapped = mapper("test.host", "testuser", name).String()

		// check that the mapped path is valid
		_, err := image.NewName(mapped)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when flattening is performed", func() {
		BeforeEach(func() {
			mapper = flattenRepoPath
		})

		Context("when the image path has a single element", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("test.host/testuser/some-user-9482d6a53a1789fb7304a4fe88362903"))
			})
		})

		Context("when the image path has more than a single element", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/some/path")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("test.host/testuser/some-user-some-path-3236c106420c1d0898246e1d2b6ba8b6"))
			})
		})

		Context("when the image has a tag", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/some/path:v1")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("test.host/testuser/some-user-some-path-93d020dc3298afd51ebcf0ce83a4ac46"))
			})
		})

		Context("when the image has a digest", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/some/path@sha256:1e725169f37aec55908694840bc808fb13ebf02cb1765df225437c56a796f870")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("test.host/testuser/some-user-some-path-bca0444f4f828ecdd7e79166b187b69e"))
			})
		})

		Context("when the image has a digest and a tag", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/some/path:v1@sha256:1e725169f37aec55908694840bc808fb13ebf02cb1765df225437c56a796f870")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("test.host/testuser/some-user-some-path-770e3a6954e3598917ef351dcd79024d"))
			})
		})

		Context("when the image path is long and has many elements", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/axxxxxxxxx/bxxxxxxxxx/cxxxxxxxxx/dxxxxxxxxx/" +
					"exxxxxxxxx/fxxxxxxxxx/gxxxxxxxxx/hxxxxxxxxx/ixxxxxxxxx/jxxxxxxxxx/kxxxxxxxxx/lxxxxxxxxx/" +
					"mxxxxxxxxx/nxxxxxxxxx/oxxxxxxxxx/pxxxxxxxxx/qxxxxxxxxx/rxxxxxxxxx/sxxxxxxxxx/txxxxxxxxx")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should omit some portions of the path", func() {
				Expect(mapped).To(Equal("test.host/testuser/some-user-axxxxxxxxx-bxxxxxxxxx-cxxxxxxxxx-dxxxxxxxxx-exxxxxxxxx-fxxxxxxxxx-gxxxxxxxxx-hxxxxxxxxx-ixxxxxxxxx-jxxxxxxxxx-kxxxxxxxxx-lxxxxxxxxx-mxxxxxxxxx-nxxxxxxxxx-oxxxxxxxxx-pxxxxxxxxx---txxxxxxxxx-b0d16e8b4d43f2ec842cfcc61989a966"))
			})
		})

		Context("when the image path is long and has few elements", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/axxxxxxxxxabxxxxxxxxxbcxxxxxxxxxcdxxxxxxxxxd" +
					"exxxxxxxxxefxxxxxxxxx/gxxxxxxxxxghxxxxxxxxxhixxxxxxxxxijxxxxxxxxxjkxxxxxxxxxklxxxxxxxxxl" +
					"mxxxxxxxxxmnxxxxxxxxxnoxxxxxxxxxopxxxxxxxxxpqxxxxxxxxxqrxxxxxxxxxrsxxxxxxxxx/txxxxxxxxx")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should omit some portions of the path", func() {
				Expect(mapped).To(Equal("test.host/testuser/some-user-axxxxxxxxxabxxxxxxxxxbcxxxxxxxxxcdxxxxxxxxxdexxxxxxxxxefxxxxxxxxx---txxxxxxxxx-4817a2fce97ff7aae687d14e66328781"))
			})
		})

		Context("when the image path is long and has three elements", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/axxxxxxxxxabxxxxxxxxxbcxxxxxxxxxcdxxxxxxxxxd" +
					"exxxxxxxxxefxxxxxxxxxfgxxxxxxxxxghxxxxxxxxxhixxxxxxxxxijxxxxxxxxxjkxxxxxxxxxklxxxxxxxxxl" +
					"mxxxxxxxxxmnxxxxxxxxxnoxxxxxxxxxopxxxxxxxxxpqxxxxxxxxxqrxxxxxxxxxrsxxxxxxxxx/txxxxxxxxx")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should omit some portions of the path", func() {
				Expect(mapped).To(Equal("test.host/testuser/some-user---txxxxxxxxx-5134b4594954926468fe0bdc23f640ef"))
			})
		})

		Context("when the image path is long and has two elements", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/axxxxxxxxxabxxxxxxxxxbcxxxxxxxxxcdxxxxxxxxxd" +
					"exxxxxxxxxefxxxxxxxxxfgxxxxxxxxxghxxxxxxxxxhixxxxxxxxxijxxxxxxxxxjkxxxxxxxxxklxxxxxxxxxl" +
					"mxxxxxxxxxmnxxxxxxxxxnoxxxxxxxxxopxxxxxxxxxpqxxxxxxxxxqrxxxxxxxxxrsxxxxxxxxxstxxxxxxxxx")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should omit some portions of the path", func() {
				Expect(mapped).To(Equal("test.host/testuser/a363dc80420c33618202ba2828aec456"))
			})
		})
	})
})
