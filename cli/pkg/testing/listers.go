/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testing

import (
	fakeprojectriffclientset "github.com/projectriff/riff/system/pkg/client/clientset/versioned/fake"
	fakeapiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakekubeclientset.AddToScheme,
	fakeapiextensionsclientset.AddToScheme,
	fakeprojectriffclientset.AddToScheme,
}

type Listers struct {
	sorter ObjectSorter
}

func NewListers(objs []runtime.Object) Listers {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		addTo(scheme)
	}

	ls := Listers{
		sorter: NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

func (l *Listers) indexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

func (l *Listers) GetAPIExtensionsObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeapiextensionsclientset.AddToScheme)
}

func (l *Listers) GetProjectriffObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeprojectriffclientset.AddToScheme)
}
