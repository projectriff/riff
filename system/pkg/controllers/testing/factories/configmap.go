/*
Copyright 2019 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package factories

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/projectriff/riff/system/pkg/apis"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
)

type configMap struct {
	target *corev1.ConfigMap
}

var (
	_ rtesting.Factory = (*configMap)(nil)
)

func ConfigMap(seed ...*corev1.ConfigMap) *configMap {
	var target *corev1.ConfigMap
	switch len(seed) {
	case 0:
		target = &corev1.ConfigMap{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &configMap{
		target: target,
	}
}

func (f *configMap) deepCopy() *configMap {
	return ConfigMap(f.target.DeepCopy())
}

func (f *configMap) Create() *corev1.ConfigMap {
	return f.deepCopy().target
}

func (f *configMap) CreateObject() apis.Object {
	return f.Create()
}

func (f *configMap) mutation(m func(*corev1.ConfigMap)) *configMap {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *configMap) NamespaceName(namespace, name string) *configMap {
	return f.mutation(func(cm *corev1.ConfigMap) {
		cm.ObjectMeta.Namespace = namespace
		cm.ObjectMeta.Name = name
	})
}

func (f *configMap) ObjectMeta(nf func(ObjectMeta)) *configMap {
	return f.mutation(func(cm *corev1.ConfigMap) {
		omf := objectMeta(cm.ObjectMeta)
		nf(omf)
		cm.ObjectMeta = omf.Create()
	})
}

func (f *configMap) AddData(key, value string) *configMap {
	return f.mutation(func(cm *corev1.ConfigMap) {
		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data[key] = value
	})
}
