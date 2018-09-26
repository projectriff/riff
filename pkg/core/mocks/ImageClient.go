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

package mocks

import core "github.com/projectriff/riff/pkg/core"
import mock "github.com/stretchr/testify/mock"

// ImageClient is an autogenerated mock type for the ImageClient type
type ImageClient struct {
	mock.Mock
}

// PullImages provides a mock function with given fields: options
func (_m *ImageClient) PullImages(options core.PullImagesOptions) error {
	ret := _m.Called(options)

	var r0 error
	if rf, ok := ret.Get(0).(func(core.PullImagesOptions) error); ok {
		r0 = rf(options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PushImages provides a mock function with given fields: options
func (_m *ImageClient) PushImages(options core.PushImagesOptions) error {
	ret := _m.Called(options)

	var r0 error
	if rf, ok := ret.Get(0).(func(core.PushImagesOptions) error); ok {
		r0 = rf(options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RelocateImages provides a mock function with given fields: options
func (_m *ImageClient) RelocateImages(options core.RelocateImagesOptions) error {
	ret := _m.Called(options)

	var r0 error
	if rf, ok := ret.Get(0).(func(core.RelocateImagesOptions) error); ok {
		r0 = rf(options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
