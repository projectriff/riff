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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/projectriff/riff/system/pkg/controllers/testing"
)

type ObjectMeta interface {
	Create() metav1.ObjectMeta

	Namespace(namespace string) ObjectMeta
	Name(format string, a ...interface{}) ObjectMeta
	GenerateName(format string, a ...interface{}) ObjectMeta
	AddLabel(key, value string) ObjectMeta
	AddAnnotation(key, value string) ObjectMeta
	Generation(generation int64) ObjectMeta
	ControlledBy(owner testing.Factory, scheme *runtime.Scheme) ObjectMeta
	Created(sec int64) ObjectMeta
	Deleted(sec int64) ObjectMeta
	UID(uid string) ObjectMeta
}

type objectMetaImpl struct {
	target *metav1.ObjectMeta
}

func objectMeta(seed metav1.ObjectMeta) *objectMetaImpl {
	return &objectMetaImpl{
		target: &seed,
	}
}

func (f *objectMetaImpl) Create() metav1.ObjectMeta {
	return *(f.target.DeepCopy())
}

func (f *objectMetaImpl) mutate(m func(*metav1.ObjectMeta)) ObjectMeta {
	m(f.target)
	return f
}

func (f *objectMetaImpl) Namespace(namespace string) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		om.Namespace = namespace
	})
}

func (f *objectMetaImpl) Name(format string, a ...interface{}) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		om.Name = fmt.Sprintf(format, a...)
	})
}

func (f *objectMetaImpl) GenerateName(format string, a ...interface{}) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		om.GenerateName = fmt.Sprintf(format, a...)
	})
}

func (f *objectMetaImpl) AddLabel(key, value string) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		if om.Labels == nil {
			om.Labels = map[string]string{}
		}
		om.Labels[key] = value
	})
}

func (f *objectMetaImpl) AddAnnotation(key, value string) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		if om.Annotations == nil {
			om.Annotations = map[string]string{}
		}
		om.Annotations[key] = value
	})
}

func (f *objectMetaImpl) Generation(generation int64) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		om.Generation = generation
	})
}

func (f *objectMetaImpl) ControlledBy(owner testing.Factory, scheme *runtime.Scheme) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		err := ctrl.SetControllerReference(owner.CreateObject(), om, scheme)
		if err != nil {
			panic(err)
		}
	})
}

func (f *objectMetaImpl) Created(sec int64) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		timestamp := metav1.Unix(sec, 0)
		om.CreationTimestamp = timestamp
	})
}

func (f *objectMetaImpl) Deleted(sec int64) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		timestamp := metav1.Unix(sec, 0)
		om.DeletionTimestamp = &timestamp
	})
}

func (f *objectMetaImpl) UID(uid string) ObjectMeta {
	return f.mutate(func(om *metav1.ObjectMeta) {
		om.UID = types.UID(uid)
	})
}
