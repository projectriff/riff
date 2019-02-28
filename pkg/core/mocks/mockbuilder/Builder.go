// Code generated by mockery v1.0.0. DO NOT EDIT.

package mockbuilder

import io "io"
import mock "github.com/stretchr/testify/mock"

// Builder is an autogenerated mock type for the Builder type
type Builder struct {
	mock.Mock
}

// Build provides a mock function with given fields: appDir, buildImage, runImage, repoName
func (_m *Builder) Build(appDir string, buildImage string, runImage string, repoName string) error {
	ret := _m.Called(appDir, buildImage, runImage, repoName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) error); ok {
		r0 = rf(appDir, buildImage, runImage, repoName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Run provides a mock function with given fields: appDir, buildImage, runImage, ports
func (_m *Builder) Run(appDir string, buildImage string, runImage string, ports []string) error {
	ret := _m.Called(appDir, buildImage, runImage, ports)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, []string) error); ok {
		r0 = rf(appDir, buildImage, runImage, ports)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetStdIo provides a mock function with given fields: out, err
func (_m *Builder) SetStdIo(out io.Writer, err io.Writer) {
	_m.Called(out, err)
}
