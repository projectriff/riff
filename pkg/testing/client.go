/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testing

import (
	projectriffclientset "github.com/projectriff/system/pkg/client/clientset/versioned/fake"
	buildv1alpha1clientset "github.com/projectriff/system/pkg/client/clientset/versioned/typed/build/v1alpha1"
	requestv1alpha1clientset "github.com/projectriff/system/pkg/client/clientset/versioned/typed/request/v1alpha1"
	streamv1alpha1clientset "github.com/projectriff/system/pkg/client/clientset/versioned/typed/stream/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetes "k8s.io/client-go/kubernetes/fake"
	corev1clientset "k8s.io/client-go/kubernetes/typed/core/v1"
)

type FakeClient struct {
	Namespace      string
	FakeKubeClient *kubernetes.Clientset
	FakeRiffClient *projectriffclientset.Clientset
}

func (c *FakeClient) DefaultNamespace() string {
	return c.Namespace
}

func (c *FakeClient) Core() corev1clientset.CoreV1Interface {
	return c.FakeKubeClient.CoreV1()
}

func (c *FakeClient) Build() buildv1alpha1clientset.BuildV1alpha1Interface {
	return c.FakeRiffClient.BuildV1alpha1()
}

func (c *FakeClient) Request() requestv1alpha1clientset.RequestV1alpha1Interface {
	return c.FakeRiffClient.RequestV1alpha1()
}

func (c *FakeClient) Stream() streamv1alpha1clientset.StreamV1alpha1Interface {
	return c.FakeRiffClient.StreamV1alpha1()
}

func NewClient(objects ...runtime.Object) *FakeClient {
	lister := NewListers(objects)

	kubeClient := kubernetes.NewSimpleClientset(lister.GetKubeObjects()...)
	riffClient := projectriffclientset.NewSimpleClientset(lister.GetProjectriffObjects()...)

	return &FakeClient{
		Namespace:      "default",
		FakeKubeClient: kubeClient,
		FakeRiffClient: riffClient,
	}
}
