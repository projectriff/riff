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
package controller

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Deployer", func() {
	var (
		d deployer
	)

	BeforeEach(func() {
		d = deployer{}
	})

	Describe("buildDeployment", func() {
		var (
			function v1.Function
		)

		BeforeEach(func() {
			function = v1.Function{}
		})

		Context("when the protocol is grpc", func() {

			BeforeEach(func() {
				function.Spec.Protocol = "grpc"
			})

			It("should set the RIFF_FUNCTION_INVOKER_PROTOCOL var to grpc", func() {
				deployment := d.buildDeployment(&function)
				mainContainer := deployment.Spec.Template.Spec.Containers[0]
				Expect(mainContainer.Env).To(ContainElement(corev1.EnvVar{
					Name:  "RIFF_FUNCTION_INVOKER_PROTOCOL",
					Value: "grpc",
				}))
			})
			It("should set the sidecar --protocol and --port arg", func() {
				deployment := d.buildDeployment(&function)
				sidecarContainer := deployment.Spec.Template.Spec.Containers[1]
				args := sidecarContainer.Args
				Expect(args[indexOf(args, "--protocol")+1]).To(Equal("grpc"))
				Expect(args[indexOf(args, "--port")+1]).To(Equal("10382"))
			})

			It("should set the HTTP_PORT var", func() {
				deployment := d.buildDeployment(&function)
				mainContainer := deployment.Spec.Template.Spec.Containers[0]
				Expect(mainContainer.Env).To(ContainElement(corev1.EnvVar{
					Name:  "HTTP_PORT",
					Value: "8080",
				}))
			})

			It("should set the GRPC_PORT var", func() {
				deployment := d.buildDeployment(&function)
				mainContainer := deployment.Spec.Template.Spec.Containers[0]
				Expect(mainContainer.Env).To(ContainElement(corev1.EnvVar{
					Name:  "GRPC_PORT",
					Value: "10382",
				}))
			})
		})

		Context("when the protocol is http", func() {

			BeforeEach(func() {
				function.Spec.Protocol = "http"
			})

			It("should set the RIFF_FUNCTION_INVOKER_PROTOCOL var to http", func() {
				deployment := d.buildDeployment(&function)
				mainContainer := deployment.Spec.Template.Spec.Containers[0]
				Expect(mainContainer.Env).To(ContainElement(corev1.EnvVar{
					Name:  "RIFF_FUNCTION_INVOKER_PROTOCOL",
					Value: "http",
				}))
			})

			It("should set the sidecar --protocol and --port arg", func() {
				deployment := d.buildDeployment(&function)
				sidecarContainer := deployment.Spec.Template.Spec.Containers[1]
				args := sidecarContainer.Args
				Expect(args[indexOf(args, "--protocol")+1]).To(Equal("http"))
				Expect(args[indexOf(args, "--port")+1]).To(Equal("8080"))
			})

			It("should set the HTTP_PORT var", func() {
				deployment := d.buildDeployment(&function)
				mainContainer := deployment.Spec.Template.Spec.Containers[0]
				Expect(mainContainer.Env).To(ContainElement(corev1.EnvVar{
					Name:  "HTTP_PORT",
					Value: "8080",
				}))
			})

			It("should set the GRPC_PORT var", func() {
				deployment := d.buildDeployment(&function)
				mainContainer := deployment.Spec.Template.Spec.Containers[0]
				Expect(mainContainer.Env).To(ContainElement(corev1.EnvVar{
					Name:  "GRPC_PORT",
					Value: "10382",
				}))
			})
		})
	})

})

func indexOf(slice []string, elem string) int {
	for index, item := range slice {
		if item == elem {
			return index
		}
	}
	return -1
}
