/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Code generated by mockery v1.0.0. DO NOT EDIT.

package vendor_mocks

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import mock "github.com/stretchr/testify/mock"
import types "k8s.io/apimachinery/pkg/types"
import v1 "k8s.io/api/core/v1"
import watch "k8s.io/apimachinery/pkg/watch"

// NamespaceInterface is an autogenerated mock type for the NamespaceInterface type
type NamespaceInterface struct {
	mock.Mock
}

// Create provides a mock function with given fields: _a0
func (_m *NamespaceInterface) Create(_a0 *v1.Namespace) (*v1.Namespace, error) {
	ret := _m.Called(_a0)

	var r0 *v1.Namespace
	if rf, ok := ret.Get(0).(func(*v1.Namespace) *v1.Namespace); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Namespace) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name, options
func (_m *NamespaceInterface) Delete(name string, options *metav1.DeleteOptions) error {
	ret := _m.Called(name, options)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *metav1.DeleteOptions) error); ok {
		r0 = rf(name, options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Finalize provides a mock function with given fields: item
func (_m *NamespaceInterface) Finalize(item *v1.Namespace) (*v1.Namespace, error) {
	ret := _m.Called(item)

	var r0 *v1.Namespace
	if rf, ok := ret.Get(0).(func(*v1.Namespace) *v1.Namespace); ok {
		r0 = rf(item)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Namespace) error); ok {
		r1 = rf(item)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: name, options
func (_m *NamespaceInterface) Get(name string, options metav1.GetOptions) (*v1.Namespace, error) {
	ret := _m.Called(name, options)

	var r0 *v1.Namespace
	if rf, ok := ret.Get(0).(func(string, metav1.GetOptions) *v1.Namespace); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, metav1.GetOptions) error); ok {
		r1 = rf(name, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: opts
func (_m *NamespaceInterface) List(opts metav1.ListOptions) (*v1.NamespaceList, error) {
	ret := _m.Called(opts)

	var r0 *v1.NamespaceList
	if rf, ok := ret.Get(0).(func(metav1.ListOptions) *v1.NamespaceList); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.NamespaceList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(metav1.ListOptions) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Patch provides a mock function with given fields: name, pt, data, subresources
func (_m *NamespaceInterface) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1.Namespace, error) {
	_va := make([]interface{}, len(subresources))
	for _i := range subresources {
		_va[_i] = subresources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, pt, data)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *v1.Namespace
	if rf, ok := ret.Get(0).(func(string, types.PatchType, []byte, ...string) *v1.Namespace); ok {
		r0 = rf(name, pt, data, subresources...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, types.PatchType, []byte, ...string) error); ok {
		r1 = rf(name, pt, data, subresources...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: _a0
func (_m *NamespaceInterface) Update(_a0 *v1.Namespace) (*v1.Namespace, error) {
	ret := _m.Called(_a0)

	var r0 *v1.Namespace
	if rf, ok := ret.Get(0).(func(*v1.Namespace) *v1.Namespace); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Namespace) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStatus provides a mock function with given fields: _a0
func (_m *NamespaceInterface) UpdateStatus(_a0 *v1.Namespace) (*v1.Namespace, error) {
	ret := _m.Called(_a0)

	var r0 *v1.Namespace
	if rf, ok := ret.Get(0).(func(*v1.Namespace) *v1.Namespace); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Namespace)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.Namespace) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Watch provides a mock function with given fields: opts
func (_m *NamespaceInterface) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	ret := _m.Called(opts)

	var r0 watch.Interface
	if rf, ok := ret.Get(0).(func(metav1.ListOptions) watch.Interface); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(watch.Interface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(metav1.ListOptions) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
