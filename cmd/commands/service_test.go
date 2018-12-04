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

package commands_test

import (
	"fmt"
	"net"
	"net/http"

	"strings"

	"time"

	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("The riff service create command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			cc         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			cc = commands.ServiceCreate(&mockClient)
		})
		It("should fail with no args", func() {
			cc.SetArgs([]string{})
			err := cc.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
		It("should fail with invalid service name", func() {
			cc.SetArgs([]string{".invalid"})
			err := cc.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
		It("should fail without required flags", func() {
			cc.SetArgs([]string{"my-service"})
			err := cc.Execute()
			Expect(err).To(MatchError(`required flag(s) "image" not set`))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			sc     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			sc = commands.ServiceCreate(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())

		})
		It("should involve the core.Client", func() {
			sc.SetArgs([]string{"my-service", "--image", "foo/bar", "--namespace", "ns"})

			o := core.CreateOrUpdateServiceOptions{
				Name:    "my-service",
				Image:   "foo/bar",
				Env:     []string{},
				EnvFrom: []string{},
			}
			o.Namespace = "ns"

			asMock.On("CreateService", o).Return(nil, nil)
			err := sc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			sc.SetArgs([]string{"my-service", "--image", "foo/bar"})

			e := fmt.Errorf("some error")
			asMock.On("CreateService", mock.Anything).Return(nil, e)
			err := sc.Execute()
			Expect(err).To(MatchError(e))
		})
		It("should add env vars when asked to", func() {
			sc.SetArgs([]string{"my-service", "--image", "foo/bar", "--namespace", "ns", "--env", "FOO=bar",
				"--env", "BAZ=qux", "--env-from", "secretKeyRef:foo:bar"})

			o := core.CreateOrUpdateServiceOptions{
				Name:    "my-service",
				Image:   "foo/bar",
				Env:     []string{"FOO=bar", "BAZ=qux"},
				EnvFrom: []string{"secretKeyRef:foo:bar"},
			}
			o.Namespace = "ns"

			asMock.On("CreateService", o).Return(nil, nil)
			err := sc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should print when --dry-run is set", func() {
			sc.SetArgs([]string{"square", "--image", "foo/bar", "--dry-run"})

			serviceOptions := core.CreateOrUpdateServiceOptions{
				Name:    "square",
				Image:   "foo/bar",
				Env:     []string{},
				EnvFrom: []string{},
				DryRun:  true,
			}

			svc := v1alpha1.Service{}
			svc.Name = "square"
			c := eventingv1alpha1.Channel{}
			c.Name = "my-channel"
			s := eventingv1alpha1.Subscription{}
			s.Name = "square"
			asMock.On("CreateService", serviceOptions).Return(&svc, nil)

			stdout := &strings.Builder{}
			sc.SetOutput(stdout)

			err := sc.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.String()).To(Equal(serviceCreateDryRun))
		})

	})
})

const serviceCreateDryRun = `metadata:
  creationTimestamp: null
  name: square
spec: {}
status: {}
---
`

var _ = Describe("The riff service update command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			cc         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			cc = commands.ServiceUpdate(&mockClient)
		})
		It("should fail with no args", func() {
			cc.SetArgs([]string{})
			err := cc.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
		It("should fail with invalid service name", func() {
			cc.SetArgs([]string{".invalid"})
			err := cc.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			sc     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			sc = commands.ServiceUpdate(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())
		})
		It("should involve the core.Client", func() {
			sc.SetArgs([]string{"my-service", "--image", "foo/bar", "--namespace", "ns"})

			o := core.CreateOrUpdateServiceOptions{
				Name:    "my-service",
				Image:   "foo/bar",
				Env:     []string{},
				EnvFrom: []string{},
			}
			o.Namespace = "ns"

			asMock.On("UpdateService", o).Return(nil, nil)
			err := sc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			sc.SetArgs([]string{"my-service", "--image", "foo/bar"})

			e := fmt.Errorf("some error")
			asMock.On("UpdateService", mock.Anything).Return(nil, e)
			err := sc.Execute()
			Expect(err).To(MatchError(e))
		})
		It("should add env vars when asked to", func() {
			sc.SetArgs([]string{"my-service", "--image", "foo/bar", "--namespace", "ns", "--env", "FOO=bar",
				"--env", "BAZ=qux", "--env-from", "secretKeyRef:foo:bar"})

			o := core.CreateOrUpdateServiceOptions{
				Name:    "my-service",
				Image:   "foo/bar",
				Env:     []string{"FOO=bar", "BAZ=qux"},
				EnvFrom: []string{"secretKeyRef:foo:bar"},
			}
			o.Namespace = "ns"

			asMock.On("UpdateService", o).Return(nil, nil)
			err := sc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should print when --dry-run is set", func() {
			sc.SetArgs([]string{"square", "--image", "foo/bar", "--dry-run"})

			serviceOptions := core.CreateOrUpdateServiceOptions{
				Name:    "square",
				Image:   "foo/bar",
				Env:     []string{},
				EnvFrom: []string{},
				DryRun:  true,
			}
			svc := v1alpha1.Service{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "square",
				},
				Spec: v1alpha1.ServiceSpec{
					RunLatest: &v1alpha1.RunLatestType{
						Configuration: v1alpha1.ConfigurationSpec{
							RevisionTemplate: v1alpha1.RevisionTemplateSpec{
								Spec: v1alpha1.RevisionSpec{
									Container: v1.Container{
										Image: "foo/bar",
									},
								},
							},
						},
					},
				},
			}
			asMock.On("UpdateService", serviceOptions).Return(&svc, nil)

			stdout := &strings.Builder{}
			sc.SetOutput(stdout)

			err := sc.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.String()).To(Equal(serviceUpdateDryRun))
		})

	})
})

const serviceUpdateDryRun = `metadata:
  creationTimestamp: null
  name: square
spec:
  runLatest:
    configuration:
      revisionTemplate:
        metadata:
          creationTimestamp: null
        spec:
          container:
            image: foo/bar
            name: ""
            resources: {}
status: {}
---
`

var _ = Describe("The riff service status command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			ss         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			ss = commands.ServiceStatus(&mockClient)
		})
		It("should fail with no args", func() {
			ss.SetArgs([]string{})
			err := ss.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
		It("should fail with invalid service name", func() {
			ss.SetArgs([]string{".invalid"})
			err := ss.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			ss     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			ss = commands.ServiceStatus(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())

		})
		It("should involve the core.Client", func() {
			ss.SetArgs([]string{"my-service", "--namespace", "ns"})

			o := core.ServiceStatusOptions{
				Name: "my-service",
			}
			o.Namespace = "ns"

			sc := &duckv1alpha1.Condition{
				Type:    v1alpha1.ServiceConditionReady,
				Status:  v1.ConditionFalse,
				Message: "punk broke",
				Reason:  "Becuz",
				LastTransitionTime: apis.VolatileTime{
					Inner: meta_v1.Date(1991, 7, 21, 19, 32, 00, 0, time.FixedZone("Europe", 0)),
				},
			}

			asMock.On("ServiceStatus", o).Return(sc, nil)

			stdout := &strings.Builder{}
			ss.SetOutput(stdout)
			err := ss.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal(svcStatusOutput))
		})
		It("should propagate core.Client errors", func() {
			ss.SetArgs([]string{"my-service"})

			e := fmt.Errorf("some error")
			asMock.On("ServiceStatus", mock.Anything).Return(nil, e)
			err := ss.Execute()
			Expect(err).To(MatchError(e))
		})
	})
})

const svcStatusOutput = `Last Transition Time:        1991-07-21T19:32:00Z
Message:                     punk broke
Reason:                      Becuz
Status:                      False
Type:                        Ready
`

var _ = Describe("The riff service list command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			sl         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			sl = commands.ServiceList(&mockClient)
		})
		It("should fail with args", func() {
			sl.SetArgs([]string{"something"})
			err := sl.Execute()
			Expect(err).To(MatchError("accepts 0 arg(s), received 1"))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			sl     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			sl = commands.ServiceList(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())

		})
		It("should involve the core.Client", func() {
			sl.SetArgs([]string{"--namespace", "ns"})

			o := core.ListServiceOptions{}
			o.Namespace = "ns"

			list := &v1alpha1.ServiceList{
				Items: []v1alpha1.Service{
					{
						ObjectMeta: meta_v1.ObjectMeta{Name: "foo"},
						Status: v1alpha1.ServiceStatus{Conditions: duckv1alpha1.Conditions{
							{
								Type:    v1alpha1.ServiceConditionReady,
								Reason:  "Failed",
								Message: "It's dead, Jim",
								Status:  v1.ConditionFalse,
							},
						}},
					},
					{
						ObjectMeta: meta_v1.ObjectMeta{Name: "wizz"},
						Status: v1alpha1.ServiceStatus{Conditions: duckv1alpha1.Conditions{
							{
								Type:   v1alpha1.ServiceConditionReady,
								Status: v1.ConditionTrue,
							},
						}},
					},
				},
			}

			asMock.On("ListServices", o).Return(list, nil)

			stdout := &strings.Builder{}
			sl.SetOutput(stdout)
			err := sl.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal(svcListOutput))
		})
		It("should propagate core.Client errors", func() {
			e := fmt.Errorf("some error")
			asMock.On("ListServices", mock.Anything).Return(nil, e)
			err := sl.Execute()
			Expect(err).To(MatchError(e))
		})
	})
})

const svcListOutput = `NAME STATUS                 
foo  Failed: It's dead, Jim 
wizz Running                

list completed successfully
`

var _ = Describe("The riff service delete command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			sd         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			sd = commands.ServiceDelete(&mockClient)
		})
		It("should fail with no args", func() {
			sd.SetArgs([]string{})
			err := sd.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
		It("should fail with invalid service name", func() {
			sd.SetArgs([]string{".invalid"})
			err := sd.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			sd     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			sd = commands.ServiceDelete(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())
		})
		It("should involve the core.Client", func() {
			sd.SetArgs([]string{"my-service", "--namespace", "ns"})

			o := core.DeleteServiceOptions{
				Name: "my-service",
			}
			o.Namespace = "ns"

			asMock.On("DeleteService", o).Return(nil)
			err := sd.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			sd.SetArgs([]string{"my-service"})

			e := fmt.Errorf("some error")
			asMock.On("DeleteService", mock.Anything).Return(e)
			err := sd.Execute()
			Expect(err).To(MatchError(e))
		})

	})
})

var _ = Describe("The riff service invoke command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient    core.Client
			invokeCommand *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			invokeCommand = commands.ServiceInvoke(&mockClient)
		})
		It("should fail with no args", func() {
			invokeCommand.SetArgs([]string{})
			err := invokeCommand.Execute()
			Expect(err).To(MatchError("requires at least 1 arg(s), only received 0"))
		})
		It("should fail with too many args", func() {
			invokeCommand.SetArgs([]string{"someservice", "/path", "oops-extra-arg"})
			err := invokeCommand.Execute()
			Expect(err).To(MatchError("accepts at most 2 arg(s), received 3"))
		})
		It("should fail with invalid service name", func() {
			invokeCommand.SetArgs([]string{".invalid"})
			err := invokeCommand.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client             core.Client
			clientMock         *mocks.Client
			invokeCommand      *cobra.Command
			listener           net.Listener
			pathMatchedChannel chan bool
			timeout            = 2 * time.Second
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			clientMock = client.(*mocks.Client)
			pathMatchedChannel = make(chan bool, 1)

			invokeCommand = commands.ServiceInvoke(&client)
		})
		It("should invoke the service", func() {
			listener = pathAwareHttpServer("/", pathMatchedChannel)
			invokeCommand.SetArgs([]string{"correlator"})
			options := core.ServiceInvokeOptions{
				Name: "correlator",
			}
			clientMock.On("ServiceCoordinates", options).Return(listener.Addr().String(), "hostname", nil)
			err := invokeCommand.Execute()

			Expect(err).To(BeNil(), "service invoke should work")
			select {
			case matchedChannel := <-pathMatchedChannel:
				Expect(matchedChannel).To(BeTrue(), "curl should reach the service")
			case <-time.After(timeout):
				Fail(fmt.Sprintf("service invoke did not complete within %v", timeout))
			}
		})
		It("should invoke the service with curl arguments", func() {
			listener = pathAwareHttpServer("/", pathMatchedChannel)
			invokeCommand.SetArgs([]string{"numbers", "--", "-HContent-Type:text/plain", "-d 7"})
			options := core.ServiceInvokeOptions{
				Name: "numbers",
			}
			clientMock.On("ServiceCoordinates", options).Return(listener.Addr().String(), "hostname", nil)
			err := invokeCommand.Execute()

			Expect(err).To(BeNil(), "service invoke should work")
			select {
			case matchedChannel := <-pathMatchedChannel:
				Expect(matchedChannel).To(BeTrue(), "curl should reach the service")
			case <-time.After(timeout):
				Fail(fmt.Sprintf("service invoke did not complete within %v", timeout))
			}
		})
		It("should accept an additional optional path argument", func() {
			path := "/numbers"
			listener = pathAwareHttpServer(path, pathMatchedChannel)
			invokeCommand.SetArgs([]string{"correlator", path})
			options := core.ServiceInvokeOptions{
				Name: "correlator",
			}
			clientMock.On("ServiceCoordinates", options).Return(listener.Addr().String(), "hostname", nil)
			err := invokeCommand.Execute()

			Expect(err).To(BeNil(), "service invoke should work with a path")
			select {
			case matchedChannel := <-pathMatchedChannel:
				Expect(matchedChannel).To(BeTrue(), "curl should take the path into account")
			case <-time.After(timeout):
				Fail(fmt.Sprintf("service invoke did not complete within %v", timeout))
			}
		})
		AfterEach(func() {
			clientMock.AssertExpectations(GinkgoT())
			if listener != nil {
				listener.Close()
			}
		})
	})
})

func pathAwareHttpServer(path string, pathMatchedChannel chan<- bool) net.Listener {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	go http.Serve(listener, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path != path {
			resp.WriteHeader(404)
			pathMatchedChannel <- false
		} else {
			resp.WriteHeader(200)
			pathMatchedChannel <- true
		}
		resp.Write([]byte{})
	}))
	return listener
}
