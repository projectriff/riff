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
	corev1alpha1clientset "github.com/projectriff/system/pkg/client/clientset/versioned/typed/core/v1alpha1"
	knativev1alpha1clientset "github.com/projectriff/system/pkg/client/clientset/versioned/typed/knative/v1alpha1"
	streamv1alpha1clientset "github.com/projectriff/system/pkg/client/clientset/versioned/typed/streaming/v1alpha1"
	apiextensionsv1beta1clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetes "k8s.io/client-go/kubernetes/fake"
	authv1client "k8s.io/client-go/kubernetes/typed/authorization/v1"
	corev1clientset "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type FakeClient struct {
	Namespace                  string
	FakeKubeRestConfig         *rest.Config
	FakeKubeClientset          *kubernetes.Clientset
	FakeRiffClientset          *projectriffclientset.Clientset
	FakeAPIExtensionsClientset *apiextensionsv1beta1clientset.Clientset
	ActionRecorderList         ActionRecorderList
}

func (c *FakeClient) DefaultNamespace() string {
	return c.Namespace
}

func (c *FakeClient) KubeRestConfig() *rest.Config {
	return c.FakeKubeRestConfig
}

func (c *FakeClient) Core() corev1clientset.CoreV1Interface {
	return c.FakeKubeClientset.CoreV1()
}

func (c *FakeClient) Auth() authv1client.AuthorizationV1Interface {
	return c.FakeKubeClientset.AuthorizationV1()
}

func (c *FakeClient) APIExtension() apiextensionsv1beta1.ApiextensionsV1beta1Interface {
	return c.FakeAPIExtensionsClientset.ApiextensionsV1beta1()
}

func (c *FakeClient) Build() buildv1alpha1clientset.BuildV1alpha1Interface {
	return c.FakeRiffClientset.BuildV1alpha1()
}

func (c *FakeClient) CoreRuntime() corev1alpha1clientset.CoreV1alpha1Interface {
	return c.FakeRiffClientset.CoreV1alpha1()
}

func (c *FakeClient) StreamingRuntime() streamv1alpha1clientset.StreamingV1alpha1Interface {
	return c.FakeRiffClientset.StreamingV1alpha1()
}

func (c *FakeClient) KnativeRuntime() knativev1alpha1clientset.KnativeV1alpha1Interface {
	return c.FakeRiffClientset.KnativeV1alpha1()
}

func (c *FakeClient) PrependReactor(verb, resource string, reaction ReactionFunc) {
	c.FakeKubeClientset.PrependReactor(verb, resource, reaction)
	c.FakeAPIExtensionsClientset.PrependReactor(verb, resource, reaction)
	c.FakeRiffClientset.PrependReactor(verb, resource, reaction)
}

func NewClient(objects ...runtime.Object) *FakeClient {
	lister := NewListers(objects)

	kubeRestConfig := &rest.Config{Host: "https://localhost:8443"}
	kubeClientset := kubernetes.NewSimpleClientset(lister.GetKubeObjects()...)
	apiExtensionsClientset := apiextensionsv1beta1clientset.NewSimpleClientset(lister.GetAPIExtensionsObjects()...)
	riffClientset := projectriffclientset.NewSimpleClientset(lister.GetProjectriffObjects()...)

	actionRecorderList := ActionRecorderList{kubeClientset, apiExtensionsClientset, riffClientset}

	return &FakeClient{
		Namespace:                  "default",
		FakeKubeRestConfig:         kubeRestConfig,
		FakeKubeClientset:          kubeClientset,
		FakeAPIExtensionsClientset: apiExtensionsClientset,
		FakeRiffClientset:          riffClientset,
		ActionRecorderList:         actionRecorderList,
	}
}
