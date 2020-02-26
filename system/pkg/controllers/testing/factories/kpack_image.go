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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectriff/system/pkg/apis"
	kpackbuildv1alpha1 "github.com/projectriff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
	rtesting "github.com/projectriff/system/pkg/controllers/testing"
)

type kpackImage struct {
	target *kpackbuildv1alpha1.Image
}

var (
	_ rtesting.Factory = (*kpackImage)(nil)
)

func KpackImage(seed ...*kpackbuildv1alpha1.Image) *kpackImage {
	var target *kpackbuildv1alpha1.Image
	switch len(seed) {
	case 0:
		target = &kpackbuildv1alpha1.Image{}
	case 1:
		target = seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &kpackImage{
		target: target,
	}
}

func (f *kpackImage) deepCopy() *kpackImage {
	return KpackImage(f.target.DeepCopy())
}

func (f *kpackImage) Create() *kpackbuildv1alpha1.Image {
	return f.deepCopy().target
}

func (f *kpackImage) CreateObject() apis.Object {
	return f.Create()
}

func (f *kpackImage) mutation(m func(*kpackbuildv1alpha1.Image)) *kpackImage {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *kpackImage) NamespaceName(namespace, name string) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.ObjectMeta.Namespace = namespace
		image.ObjectMeta.Name = name
	})
}

func (f *kpackImage) ObjectMeta(nf func(ObjectMeta)) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		omf := objectMeta(image.ObjectMeta)
		nf(omf)
		image.ObjectMeta = omf.Create()
	})
}

func (f *kpackImage) ApplicationBuilder() *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.Spec.Builder = kpackbuildv1alpha1.ImageBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind: "ClusterBuilder",
			},
			Name: "riff-application",
		}
		image.Spec.ServiceAccount = "riff-build"
	})
}

func (f *kpackImage) FunctionBuilder(artifact, handler, invoker string) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.Spec.Builder = kpackbuildv1alpha1.ImageBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind: "ClusterBuilder",
			},
			Name: "riff-function",
		}
		image.Spec.ServiceAccount = "riff-build"
		env := []corev1.EnvVar{}
		for _, envvar := range image.Spec.Build.Env {
			// filter existing value
			if envvar.Name != "RIFF" && envvar.Name != "RIFF_ARTIFACT" && envvar.Name != "RIFF_HANDLER" && envvar.Name != "RIFF_OVERRIDE" {
				env = append(env, envvar)
			}
		}
		// add new values
		image.Spec.Build.Env = append(env,
			corev1.EnvVar{
				Name:  "RIFF",
				Value: "true",
			},
			corev1.EnvVar{
				Name:  "RIFF_ARTIFACT",
				Value: artifact,
			},
			corev1.EnvVar{
				Name:  "RIFF_HANDLER",
				Value: handler,
			},
			corev1.EnvVar{
				Name:  "RIFF_OVERRIDE",
				Value: invoker,
			},
		)
	})
}

func (f *kpackImage) Tag(format string, a ...interface{}) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.Spec.Tag = fmt.Sprintf(format, a...)
	})
}

func (f *kpackImage) SourceGit(url string, revision string) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.Spec.Source = kpackbuildv1alpha1.SourceConfig{
			Git: &kpackbuildv1alpha1.Git{
				URL:      url,
				Revision: revision,
			},
			SubPath: image.Spec.Source.SubPath,
		}
	})
}

func (f *kpackImage) SourceSubPath(subpath string) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.Spec.Source.SubPath = subpath
	})
}

func (f *kpackImage) BuildCache(quantity string) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		size, err := resource.ParseQuantity(quantity)
		if err != nil {
			panic(err)
		}
		image.Spec.CacheSize = &size
	})
}

func (f *kpackImage) StatusConditions(conditions ...*condition) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		c := make([]apis.Condition, len(conditions))
		for i, cg := range conditions {
			c[i] = cg.Create()
		}
		image.Status.Conditions = c
	})
}

func (f *kpackImage) StatusReady() *kpackImage {
	return f.StatusConditions(
		Condition().Type(apis.ConditionReady).True(),
	)
}

func (f *kpackImage) StatusObservedGeneration(generation int64) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.Status.ObservedGeneration = generation
	})
}

func (f *kpackImage) StatusLatestImage(format string, a ...interface{}) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.Status.LatestImage = fmt.Sprintf(format, a...)
	})
}

func (f *kpackImage) StatusBuildCacheName(format string, a ...interface{}) *kpackImage {
	return f.mutation(func(image *kpackbuildv1alpha1.Image) {
		image.Status.BuildCacheName = fmt.Sprintf(format, a...)
	})
}
