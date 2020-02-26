/*
Copyright 2020 the original author or authors.

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

	"github.com/projectriff/system/pkg/apis"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
)

type stream struct {
	target *streamingv1alpha1.Stream
}

var (
	_ rtesting.Factory = (*stream)(nil)
)

func Stream(seed ...*streamingv1alpha1.Stream) *stream {
	var target *streamingv1alpha1.Stream
	switch len(seed) {
	case 0:
		target = &streamingv1alpha1.Stream{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &stream{
		target: target,
	}
}

func (f *stream) deepCopy() *stream {
	return Stream(f.target.DeepCopy())
}

func (f *stream) Create() *streamingv1alpha1.Stream {
	return f.deepCopy().target
}

func (f *stream) CreateObject() apis.Object {
	return f.Create()
}

func (f *stream) CreateInputStreamBinding(alias, startOffset string) streamingv1alpha1.InputStreamBinding {
	return streamingv1alpha1.InputStreamBinding{
		Stream:      f.target.Name,
		Alias:       alias,
		StartOffset: startOffset,
	}
}

func (f *stream) CreateOutputStreamBinding(alias string) streamingv1alpha1.OutputStreamBinding {
	return streamingv1alpha1.OutputStreamBinding{
		Stream: f.target.Name,
		Alias:  alias,
	}
}

func (f *stream) mutation(m func(*streamingv1alpha1.Stream)) *stream {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *stream) NamespaceName(namespace, name string) *stream {
	return f.mutation(func(s *streamingv1alpha1.Stream) {
		s.ObjectMeta.Namespace = namespace
		s.ObjectMeta.Name = name
	})
}

func (f *stream) ObjectMeta(nf func(ObjectMeta)) *stream {
	return f.mutation(func(s *streamingv1alpha1.Stream) {
		omf := objectMeta(s.ObjectMeta)
		nf(omf)
		s.ObjectMeta = omf.Create()
	})
}

func (f *stream) ContentType(mime string) *stream {
	return f.mutation(func(s *streamingv1alpha1.Stream) {
		s.Spec.ContentType = mime
	})
}

func (f *stream) Gateway(name string) *stream {
	return f.mutation(func(s *streamingv1alpha1.Stream) {
		s.Spec.Gateway.Name = name
	})
}

func (f *stream) StatusConditions(conditions ...*condition) *stream {
	return f.mutation(func(s *streamingv1alpha1.Stream) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		s.Status.Conditions = c
	})
}

func (f *stream) StatusReady() *stream {
	return f.StatusConditions(
		Condition().Type(streamingv1alpha1.StreamConditionReady).True(),
	)
}

func (f *stream) StatusObservedGeneration(generation int64) *stream {
	return f.mutation(func(s *streamingv1alpha1.Stream) {
		s.Status.ObservedGeneration = generation
	})
}

func (f *stream) StatusBinding(metadataName, secretName string) *stream {
	return f.mutation(func(s *streamingv1alpha1.Stream) {
		s.Status.Binding.MetadataRef = corev1.LocalObjectReference{Name: metadataName}
		s.Status.Binding.SecretRef = corev1.LocalObjectReference{Name: secretName}
	})
}
