/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fake

import (
	v1 "github.com/projectriff/kubernetes-crds/pkg/client/clientset/versioned/typed/projectriff/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeProjectriffV1 struct {
	*testing.Fake
}

func (c *FakeProjectriffV1) Functions(namespace string) v1.FunctionInterface {
	return &FakeFunctions{c, namespace}
}

func (c *FakeProjectriffV1) Topics(namespace string) v1.TopicInterface {
	return &FakeTopics{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeProjectriffV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
