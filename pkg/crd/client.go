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
	"github.com/projectriff/riff/pkg/env"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type riffClient struct {
	restClient rest.Interface
}
type Client interface {
	Create(obj *Manifest) (*Manifest, error)
	Update(obj *Manifest) (*Manifest, error)
	Delete(obj *Manifest) (*Manifest, error)
	Get() (*Manifest, error)
}

func NewClient(config clientcmd.ClientConfig) (Client, error) {
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
	return &riffClient{restClient: rc}, nil
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
	scheme.AddKnownTypes(schemeGroupVersion, &Manifest{})
	metav1.AddToGroupVersion(scheme, schemeGroupVersion)
	return nil
}

func (client *riffClient) Create(obj *Manifest) (*Manifest, error) {
	result := &Manifest{}
	err := client.restClient.Post().Resource(Name).Name(env.Cli.Name + "-install").
		Body(obj).Do().Into(result)
	return result, err
}

func (client *riffClient) Update(obj *Manifest) (*Manifest, error) {
	panic("not implemented")
}

func (client *riffClient) Delete(obj *Manifest) (*Manifest, error) {
	panic("not implemented")
}

func (client *riffClient) Get() (*Manifest, error) {
	result := &Manifest{}
	err := client.restClient.Get().Resource(Name).Name(env.Cli.Name + "-install").Do().Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
