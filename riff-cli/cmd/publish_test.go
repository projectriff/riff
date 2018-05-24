/*
 * Copyright 2017 the original author or authors.
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

package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/minikube"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

var _ = Describe("The publish command", func() {
	var (
		kubeClient     *kubectl.MockKubeCtl
		minik          *minikube.MockMinikube
		publishCommand *cobra.Command

		cannedNodePortReply                 string
		cannedLoadBalancerReply             string
		cannedLoadBalancerWithHostnameReply string
		server                              *ghttp.Server
	)

	BeforeEach(func() {
		var err error
		kubeClient = new(kubectl.MockKubeCtl)
		minik = new(minikube.MockMinikube)

		publishCommand = Publish(kubeClient, minik)

		server = ghttp.NewServer()
		port := strings.Split(server.Addr(), ":")[1]

		b, err := ioutil.ReadFile("../test_data/publish/NodePortReply.json")
		Expect(err).NotTo(HaveOccurred())
		cannedNodePortReply = strings.Replace(string(b), "<port>", port, 1)
		b, err = ioutil.ReadFile("../test_data/publish/LoadBalancerReply.json")
		Expect(err).NotTo(HaveOccurred())
		cannedLoadBalancerReply = strings.Replace(string(b), "<port>", port, 1)
		cannedLoadBalancerReply = strings.Replace(cannedLoadBalancerReply, "<ip-address>", "localhost", 1)
		b, err = ioutil.ReadFile("../test_data/publish/LoadBalancerWithHostnameReply.json")
		Expect(err).NotTo(HaveOccurred())
		cannedLoadBalancerWithHostnameReply = strings.Replace(string(b), "<port>", port, 1)
		cannedLoadBalancerWithHostnameReply = strings.Replace(cannedLoadBalancerWithHostnameReply, "<hostname>", "localhost", 1)
	})

	AfterEach(func() {
		kubeClient.AssertExpectations(GinkgoT())
		minik.AssertExpectations(GinkgoT())

		server.Close()

	})

	It("should use current working directory as implicit function name (and hence input topic name)", func() {
		kubeClient.
			On("Exec", []string{"get", "svc", "--all-namespaces", "-l", "app=riff,component=http-gateway", "-o", "json"}).
			Return(cannedNodePortReply, nil)
		minik.On("QueryIp").Return("localhost", nil)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/messages/cmd"),
				ghttp.VerifyBody([]byte("hello")),
				ghttp.VerifyHeaderKV("Content-Type", "text/plain"),
			),
		)

		publishCommand.SetArgs([]string{"-d", "hello"})

		err := publishCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(server.ReceivedRequests()).To(HaveLen(1))
	})

	It("should pass the provided Content-Type header", func() {
		kubeClient.
			On("Exec", []string{"get", "svc", "--all-namespaces", "-l", "app=riff,component=http-gateway", "-o", "json"}).
			Return(cannedNodePortReply, nil)
		minik.On("QueryIp").Return("localhost", nil)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/messages/cmd"),
				ghttp.VerifyBody([]byte("hello")),
				ghttp.VerifyHeaderKV("Content-Type", "animal/rabbit"),
			),
		)

		publishCommand.SetArgs([]string{"-d", "hello", "--content-type", "animal/rabbit"})

		err := publishCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(server.ReceivedRequests()).To(HaveLen(1))
	})

	It("should work with LoadBalancer service", func() {
		kubeClient.
			On("Exec", []string{"get", "svc", "--all-namespaces", "-l", "app=riff,component=http-gateway", "-o", "json"}).
			Return(cannedLoadBalancerReply, nil)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/messages/cmd"),
				ghttp.VerifyBody([]byte("hello")),
			),
		)

		publishCommand.SetArgs([]string{"-d", "hello"})

		err := publishCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(server.ReceivedRequests()).To(HaveLen(1))
	})

	It("should work with LoadBalancer service with hostname", func() {
		kubeClient.
			On("Exec", []string{"get", "svc", "--all-namespaces", "-l", "app=riff,component=http-gateway", "-o", "json"}).
			Return(cannedLoadBalancerWithHostnameReply, nil)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/messages/cmd"),
				ghttp.VerifyBody([]byte("hello")),
			),
		)

		publishCommand.SetArgs([]string{"-d", "hello"})

		err := publishCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(server.ReceivedRequests()).To(HaveLen(1))
	})

	It("should post to path based on input topic", func() {
		kubeClient.
			On("Exec", []string{"get", "svc", "--all-namespaces", "-l", "app=riff,component=http-gateway", "-o", "json"}).
			Return(cannedNodePortReply, nil)
		minik.On("QueryIp").Return("localhost", nil)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/messages/foobar"),
				ghttp.VerifyBody([]byte("hello")),
			),
		)

		publishCommand.SetArgs([]string{"-d", "hello", "--input", "foobar"})

		err := publishCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(server.ReceivedRequests()).To(HaveLen(1))
	})

	It("should use requests path when expecting a reply", func() {
		kubeClient.
			On("Exec", []string{"get", "svc", "--all-namespaces", "-l", "app=riff,component=http-gateway", "-o", "json"}).
			Return(cannedNodePortReply, nil)
		minik.On("QueryIp").Return("localhost", nil)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/requests/foobar"),
				ghttp.VerifyBody([]byte("hello")),
				ghttp.RespondWith(200, "world"),
			),
		)

		publishCommand.SetArgs([]string{"-d", "hello", "--input", "foobar", "-r"})

		err := publishCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(server.ReceivedRequests()).To(HaveLen(1))
	})

	It("should post 'count' many number of times", func() {
		kubeClient.
			On("Exec", []string{"get", "svc", "--all-namespaces", "-l", "app=riff,component=http-gateway", "-o", "json"}).
			Return(cannedNodePortReply, nil)
		minik.On("QueryIp").Return("localhost", nil)

		for i := 0; i < 3; i++ {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/messages/foobar"),
					ghttp.VerifyBody([]byte("hello")),
				),
			)
		}

		publishCommand.SetArgs([]string{"-d", "hello", "--input", "foobar", "--count", "3"})

		err := publishCommand.Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(server.ReceivedRequests()).To(HaveLen(3))
	})
})
