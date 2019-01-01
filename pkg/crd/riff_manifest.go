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

package crd

import (
	"fmt"
	"github.com/jinzhu/copier"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ResourceChecks struct {
	Kind       string               `json:"kind,omitempty"`
	Namespace  string               `json:"namespace,omitempty"`
	Selector   metav1.LabelSelector `json:"selector,omitempty"`
	JsonPath   string               `json:"jsonpath,omitempty"`
	Pattern    string               `json:"pattern,omitempty"`
}

type RiffResources struct {
	Path       string           `json:"path,omitempty"`
	Name       string           `json:"name,omitempty"`
	Checks     []ResourceChecks `json:"checks,omitempty"`
}

type RiffSpec struct {
	Images    []string        `json:"images,omitempty"`
	Resources []RiffResources `json:"resources,omitempty"`
}

type RiffStatus struct {
	Status string `json:"status,omitempty"`
}

type RiffManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RiffSpec   `json:"spec,omitempty"`
	Status RiffStatus `json:"status,omitempty"`
}

func (orig *RiffManifest) DeepCopyObject() runtime.Object {
	result := &RiffManifest{}
	copier.Copy(result, orig)
	return result
}

type riffCRDClient struct {
	restClient rest.Interface
}

var schemeGroupVersion = schema.GroupVersion{
	Group:    Group,
	Version:  Version,
}

type Client interface {
	Create(obj *RiffManifest) (*RiffManifest, error)
	Update(obj * RiffManifest) (*RiffManifest, error)
	Delete(obj * RiffManifest) (*RiffManifest, error)
	Get() (*RiffManifest, error)
}

func NewRiffCRDClient(config clientcmd.ClientConfig) (Client, error) {
	cfg, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}
	cfg.GroupVersion = &schemeGroupVersion
	cfg.APIPath = "/apis"
	cfg.ContentType = runtime.ContentTypeJSON
	cfg.NegotiatedSerializer = serializer.DirectCodecFactory{
		CodecFactory: serializer.NewCodecFactory(getScheme()),
	}
	rc, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, err
	}
	return &riffCRDClient{restClient: rc}, nil
}

func getScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		fmt.Errorf("could not create custom CRD %s", err)
		return nil
	}
	return scheme
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(schemeGroupVersion, &RiffManifest{})
	metav1.AddToGroupVersion(scheme, schemeGroupVersion)
	return nil
}

func (client *riffCRDClient) Create(obj *RiffManifest) (*RiffManifest, error) {
	result := &RiffManifest{}
	err := client.restClient.Post().Resource("riff-system").Name("riff-install").
		Body(obj).Do().Into(result)
	return result, err
}

func (client *riffCRDClient) Update(obj *RiffManifest) (*RiffManifest, error) {
	panic("not implemented")
}

func (client *riffCRDClient) Delete(obj *RiffManifest) (*RiffManifest, error) {
	panic("not implemented")
}

func (client *riffCRDClient) Get() (*RiffManifest, error) {
	result := &RiffManifest{}
	err := client.restClient.Get().Resource("riff-system").Name("riff-install").Do().Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}