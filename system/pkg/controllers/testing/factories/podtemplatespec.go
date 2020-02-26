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
	corev1 "k8s.io/api/core/v1"
)

type PodTemplateSpec interface {
	Create() corev1.PodTemplateSpec

	AddLabel(key, value string) PodTemplateSpec
	AddAnnotation(key, value string) PodTemplateSpec
	ContainerNamed(name string, cb func(*corev1.Container)) PodTemplateSpec
	Volumes(volumes ...corev1.Volume) PodTemplateSpec
}

type podTemplateSpecImpl struct {
	target *corev1.PodTemplateSpec
}

func podTemplateSpec(seed corev1.PodTemplateSpec) *podTemplateSpecImpl {
	return &podTemplateSpecImpl{
		target: &seed,
	}
}

func (f *podTemplateSpecImpl) Create() corev1.PodTemplateSpec {
	return *(f.target.DeepCopy())
}

func (f *podTemplateSpecImpl) mutate(m func(*corev1.PodTemplateSpec)) PodTemplateSpec {
	m(f.target)
	return f
}

func (f *podTemplateSpecImpl) AddLabel(key, value string) PodTemplateSpec {
	return f.mutate(func(pts *corev1.PodTemplateSpec) {
		if pts.Labels == nil {
			pts.Labels = map[string]string{}
		}
		pts.Labels[key] = value
	})
}

func (f *podTemplateSpecImpl) AddAnnotation(key, value string) PodTemplateSpec {
	return f.mutate(func(pts *corev1.PodTemplateSpec) {
		if pts.Annotations == nil {
			pts.Annotations = map[string]string{}
		}
		pts.Annotations[key] = value
	})
}

func (f *podTemplateSpecImpl) ContainerNamed(name string, cb func(*corev1.Container)) PodTemplateSpec {
	return f.mutate(func(pts *corev1.PodTemplateSpec) {
		found := false
		// check for existing container
		for i, container := range pts.Spec.Containers {
			if container.Name == name {
				found = true
				if cb != nil {
					// container mutations
					cb(&container)
					pts.Spec.Containers[i] = container
				}
				break
			}
		}
		if !found {
			// not found, create new container
			container := corev1.Container{Name: name}
			if cb != nil {
				// container mutations
				cb(&container)
			}
			pts.Spec.Containers = append(pts.Spec.Containers, container)
		}
	})
}

func (f *podTemplateSpecImpl) Volumes(volumes ...corev1.Volume) PodTemplateSpec {
	return f.mutate(func(pts *corev1.PodTemplateSpec) {
		pts.Spec.Volumes = volumes
	})
}
