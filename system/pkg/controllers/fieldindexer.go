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

package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/projectriff/riff/system/pkg/apis"
)

func IndexControllersOfType(mgr ctrl.Manager, field string, owner, ownee apis.Object, scheme *runtime.Scheme) error {
	gvks, _, err := scheme.ObjectKinds(owner)
	if err != nil {
		return err
	}
	ownerAPIVersion, ownerKind := gvks[0].ToAPIVersionAndKind()

	return mgr.GetFieldIndexer().IndexField(ownee, field, func(rawObj runtime.Object) []string {
		ownerRef := metav1.GetControllerOf(rawObj.(metav1.Object))
		if ownerRef == nil {
			return nil
		}
		if ownerRef.APIVersion != ownerAPIVersion || ownerRef.Kind != ownerKind {
			return nil
		}
		return []string{ownerRef.Name}
	})
}
