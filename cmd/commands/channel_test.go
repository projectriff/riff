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
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	eventing "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("The riff channel create command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			cc         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			cc = commands.ChannelCreate(&mockClient)
		})
		It("should fail with no args", func() {
			cc.SetArgs([]string{})
			err := cc.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
		It("should fail with invalid channel name", func() {
			cc.SetArgs([]string{".invalid"})
			err := cc.Execute()
			Expect(err).To(MatchError(ContainSubstring("must start and end with an alphanumeric character")))
		})
		It("should fail without required flags", func() {
			cc.SetArgs([]string{"my-channel"})
			err := cc.Execute()
			Expect(err).To(MatchError("at least one of --cluster-provisioner must be set"))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			cc     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			cc = commands.ChannelCreate(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())

		})
		It("should involve the core.Client", func() {
			cc.SetArgs([]string{"my-channel", "--cluster-provisioner", "ccp", "--namespace", "ns"})

			o := core.CreateChannelOptions{
				Name:                      "my-channel",
				ClusterChannelProvisioner: "ccp",
			}
			o.Namespace = "ns"

			asMock.On("CreateChannel", o).Return(nil, nil)
			err := cc.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			cc.SetArgs([]string{"my-channel", "--cluster-provisioner", "ccp", "--namespace", "ns"})

			e := fmt.Errorf("some error")
			asMock.On("CreateChannel", mock.Anything).Return(nil, e)
			err := cc.Execute()
			Expect(err).To(MatchError(e))
		})
		It("should print when --dry-run is set", func() {
			cc.SetArgs([]string{"my-channel", "--cluster-provisioner", "ccp", "--namespace", "ns", "--dry-run"})

			o := core.CreateChannelOptions{
				Name:                      "my-channel",
				ClusterChannelProvisioner: "ccp",
				DryRun:                    true,
			}
			o.Namespace = "ns"

			c := eventing.Channel{}
			c.Name = "my-channel"
			c.Spec.Provisioner = &v1.ObjectReference{
				APIVersion: "eventing.knative.dev/v1alpha1",
				Kind:       "ClusterChannelProvisioner",
				Name:       "ccp",
			}
			asMock.On("CreateChannel", o).Return(&c, nil)

			stdout := &strings.Builder{}
			cc.SetOutput(stdout)

			err := cc.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.String()).To(Equal(channelCreateDryRun))
		})

	})
})

const channelCreateDryRun = `metadata:
  creationTimestamp: null
  name: my-channel
spec:
  provisioner:
    apiVersion: eventing.knative.dev/v1alpha1
    kind: ClusterChannelProvisioner
    name: ccp
status:
  address: {}
---
`

var _ = Describe("The riff channel list command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			cl         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			cl = commands.ChannelList(&mockClient)
		})
		It("should fail with args", func() {
			cl.SetArgs([]string{"something"})
			err := cl.Execute()
			Expect(err).To(MatchError("accepts 0 arg(s), received 1"))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			cl     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			cl = commands.ChannelList(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())
		})
		It("should involve the core.Client", func() {
			cl.SetArgs([]string{"--namespace", "ns"})

			o := core.ListChannelOptions{}
			o.Namespace = "ns"

			list := &eventing.ChannelList{
				Items: []eventing.Channel{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "foo"},
						Spec: eventing.ChannelSpec{
							Provisioner: &v1.ObjectReference{APIVersion: "eventing.knative.dev/v1alpha1", Kind: "ClusterChannelProvisioner", Name: "pubsub"},
						},
						Status: eventing.ChannelStatus{
							Conditions: duckv1alpha1.Conditions{
								{Type: v1alpha1.ChannelConditionReady, Status: v1.ConditionTrue},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "bar"},
						Spec: eventing.ChannelSpec{
							Provisioner: &v1.ObjectReference{APIVersion: "eventing.knative.dev/v1alpha1", Kind: "ClusterChannelProvisioner", Name: "kafka"},
						},
						Status: eventing.ChannelStatus{
							Conditions: duckv1alpha1.Conditions{
								{Type: v1alpha1.ChannelConditionReady, Status: v1.ConditionFalse, Reason: "RevisionFailed", Message: "oopsie"},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "baz"},
						Spec: eventing.ChannelSpec{
							Provisioner: &v1.ObjectReference{APIVersion: "eventing.knative.dev/v1alpha1", Kind: "ClusterChannelProvisioner", Name: "stub"},
						},
						Status: eventing.ChannelStatus{
							Conditions: duckv1alpha1.Conditions{
								{Type: v1alpha1.ChannelConditionReady, Status: v1.ConditionUnknown},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "foobar"},
						Spec: eventing.ChannelSpec{
							Provisioner: &v1.ObjectReference{APIVersion: "eventing.knative.dev/v1alpha1", Kind: "ClusterChannelProvisioner", Name: "stub"},
						},
					},
				},
			}

			asMock.On("ListChannels", o).Return(list, nil)

			stdout := &strings.Builder{}
			cl.SetOutput(stdout)

			err := cl.Execute()
			Expect(stdout.String()).To(Equal(channelListOutput))
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			cl.SetArgs([]string{})

			e := fmt.Errorf("some error")
			asMock.On("ListChannels", mock.Anything).Return(nil, e)
			err := cl.Execute()
			Expect(err).To(MatchError(e))
		})
	})
})

const channelListOutput = `NAME   STATUS                 PROVISIONER    
foo    Running                cluster:pubsub 
bar    RevisionFailed: oopsie cluster:kafka  
baz    Unknown                cluster:stub   
foobar Unknown                cluster:stub   

list completed successfully
`

var _ = Describe("The riff channel delete command", func() {
	Context("when given wrong args or flags", func() {
		var (
			mockClient core.Client
			cd         *cobra.Command
		)
		BeforeEach(func() {
			mockClient = nil
			cd = commands.ChannelDelete(&mockClient)
		})
		It("should fail with no args", func() {
			cd.SetArgs([]string{})
			err := cd.Execute()
			Expect(err).To(MatchError("accepts 1 arg(s), received 0"))
		})
	})

	Context("when given suitable args and flags", func() {
		var (
			client core.Client
			asMock *mocks.Client
			cd     *cobra.Command
		)
		BeforeEach(func() {
			client = new(mocks.Client)
			asMock = client.(*mocks.Client)

			cd = commands.ChannelDelete(&client)
		})
		AfterEach(func() {
			asMock.AssertExpectations(GinkgoT())
		})
		It("should involve the core.Client", func() {
			cd.SetArgs([]string{"my-channel", "--namespace", "ns"})

			o := core.DeleteChannelOptions{
				Name: "my-channel",
			}
			o.Namespace = "ns"

			asMock.On("DeleteChannel", o).Return(nil, nil)

			err := cd.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
		It("should propagate core.Client errors", func() {
			cd.SetArgs([]string{"my-channel"})

			e := fmt.Errorf("some error")
			asMock.On("DeleteChannel", mock.Anything).Return(e)
			err := cd.Execute()
			Expect(err).To(MatchError(e))
		})
	})
})
