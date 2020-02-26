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

package refs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/projectriff/system/pkg/apis"
)

func NewTypedLocalObjectReference(name string, gk schema.GroupKind) *TypedLocalObjectReference {
	if name == "" || gk.Empty() {
		return nil
	}

	ref := &TypedLocalObjectReference{
		Kind: gk.Kind,
		Name: name,
	}
	if gk.Group != "" && gk.Group != "core" {
		ref.APIGroup = &gk.Group
	}
	return ref
}

func NewTypedLocalObjectReferenceForObject(obj apis.Object, scheme *runtime.Scheme) *TypedLocalObjectReference {
	if obj == nil {
		return nil
	}

	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil || len(gvks) == 0 {
		panic(fmt.Errorf("Unregistered runtime object: %v", err))
	}
	return NewTypedLocalObjectReference(obj.GetName(), gvks[0].GroupKind())
}
