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
		mapped = mapper(name.Path(), "testuser", name)
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
				Expect(mapped).To(Equal("testuser/some-user-9482d6a53a1789fb7304a4fe88362903"))
			})
		})

		Context("when the image path has more than a single element", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/some/path")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("testuser/path-3236c106420c1d0898246e1d2b6ba8b6"))
			})
		})

		Context("when the image has a tag", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/some/path:v1")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("testuser/path-93d020dc3298afd51ebcf0ce83a4ac46"))
			})
		})

		Context("when the image has a digest", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/some/path@sha256:1e725169f37aec55908694840bc808fb13ebf02cb1765df225437c56a796f870")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("testuser/path-bca0444f4f828ecdd7e79166b187b69e"))
			})
		})

		Context("when the image has a digest and a tag", func() {
			BeforeEach(func() {
				var err error
				name, err = image.NewName("some.registry.com/some-user/some/path:v1@sha256:1e725169f37aec55908694840bc808fb13ebf02cb1765df225437c56a796f870")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should flatten the path correctly", func() {
				Expect(mapped).To(Equal("testuser/path-770e3a6954e3598917ef351dcd79024d"))
			})
		})
	})
})
