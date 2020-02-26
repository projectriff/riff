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

package build

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	kpackbuildv1alpha1 "github.com/projectriff/system/pkg/apis/thirdparty/kpack/build/v1alpha1"
	"github.com/projectriff/system/pkg/controllers"
	"github.com/projectriff/system/pkg/refs"
)

// +kubebuilder:rbac:groups=build.projectriff.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.projectriff.io,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=build.pivotal.io,resources=images,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func ApplicationReconciler(c controllers.Config) *controllers.ParentReconciler {
	c.Log = c.Log.WithName("Application")

	return &controllers.ParentReconciler{
		Type: &buildv1alpha1.Application{},
		SubReconcilers: []controllers.SubReconciler{
			ApplicationTargetImageReconciler(c),
			ApplicationChildImageReconciler(c),
		},

		Config: c,
	}
}

func ApplicationTargetImageReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("TargetImage")

	return &controllers.SyncReconciler{
		Sync: func(ctx context.Context, parent *buildv1alpha1.Application) error {
			targetImage, err := resolveTargetImage(ctx, c.Client, parent)
			if err != nil {
				if err == errMissingDefaultPrefix {
					parent.Status.MarkImageDefaultPrefixMissing(err.Error())
				} else {
					parent.Status.MarkImageInvalid(err.Error())
				}
				return err
			}
			parent.Status.MarkImageResolved()
			parent.Status.TargetImage = targetImage
			return nil
		},

		Config: c,
	}
}

func ApplicationChildImageReconciler(c controllers.Config) controllers.SubReconciler {
	c.Log = c.Log.WithName("ChildImage")

	return &controllers.ChildReconciler{
		ParentType:    &buildv1alpha1.Application{},
		ChildType:     &kpackbuildv1alpha1.Image{},
		ChildListType: &kpackbuildv1alpha1.ImageList{},

		DesiredChild: func(parent *buildv1alpha1.Application) (*kpackbuildv1alpha1.Image, error) {
			if parent.Spec.Source == nil {
				return nil, nil
			}

			image := &kpackbuildv1alpha1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Labels: controllers.MergeMaps(parent.Labels, map[string]string{
						buildv1alpha1.ApplicationLabelKey: parent.Name,
					}),
					Annotations:  make(map[string]string),
					GenerateName: fmt.Sprintf("%s-application-", parent.Name),
					Namespace:    parent.Namespace,
				},
				Spec: kpackbuildv1alpha1.ImageSpec{
					Tag: parent.Status.TargetImage,
					Builder: kpackbuildv1alpha1.ImageBuilder{
						TypeMeta: metav1.TypeMeta{
							Kind: "ClusterBuilder",
						},
						Name: "riff-application",
					},
					ServiceAccount:           riffBuildServiceAccount,
					Source:                   *parent.Spec.Source,
					CacheSize:                parent.Spec.CacheSize,
					FailedBuildHistoryLimit:  parent.Spec.FailedBuildHistoryLimit,
					SuccessBuildHistoryLimit: parent.Spec.SuccessBuildHistoryLimit,
					ImageTaggingStrategy:     parent.Spec.ImageTaggingStrategy,
					Build:                    parent.Spec.Build,
				},
			}

			return image, nil
		},
		ReflectChildStatusOnParent: func(parent *buildv1alpha1.Application, child *kpackbuildv1alpha1.Image, err error) {
			if child == nil {
				// TODO resolve to a digest?
				parent.Status.LatestImage = parent.Status.TargetImage
				parent.Status.MarkBuildNotUsed()
			} else {
				parent.Status.KpackImageRef = refs.NewTypedLocalObjectReferenceForObject(child, c.Scheme)
				parent.Status.LatestImage = child.Status.LatestImage
				parent.Status.BuildCacheRef = refs.NewTypedLocalObjectReference(child.Status.BuildCacheName, schema.GroupKind{Kind: "PersistentVolumeClaim"})
				parent.Status.PropagateKpackImageStatus(&child.Status)
			}
		},
		MergeBeforeUpdate: func(current, desired *kpackbuildv1alpha1.Image) {
			current.Labels = desired.Labels
			current.Spec = desired.Spec
		},
		SemanticEquals: func(a1, a2 *kpackbuildv1alpha1.Image) bool {
			return equality.Semantic.DeepEqual(a1.Spec, a2.Spec) &&
				equality.Semantic.DeepEqual(a1.Labels, a2.Labels)
		},

		Config:     c,
		IndexField: ".metadata.applicationController",
		Sanitize: func(child *kpackbuildv1alpha1.Image) interface{} {
			return child.Spec
		},
	}
}
