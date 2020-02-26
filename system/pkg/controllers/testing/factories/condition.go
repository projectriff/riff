/*
Copyright 2020 the original author or authors.

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

	"github.com/projectriff/system/pkg/apis"
)

type condition struct {
	target *apis.Condition
}

func Condition(seed ...apis.Condition) *condition {
	var target *apis.Condition
	switch len(seed) {
	case 0:
		target = &apis.Condition{}
	case 1:
		target = &seed[0]
	default:
		panic(fmt.Errorf("expected exactly zero or one seed, got %v", seed))
	}
	return &condition{
		target: target,
	}
}

func (f *condition) deepCopy() *condition {
	return Condition(*f.target.DeepCopy())
}

func (f *condition) Create() apis.Condition {
	return *f.deepCopy().target
}

func (f *condition) mutation(m func(*apis.Condition)) *condition {
	f = f.deepCopy()
	m(f.target)
	return f
}

func (f *condition) Type(t apis.ConditionType) *condition {
	return f.mutation(func(c *apis.Condition) {
		c.Type = t
	})
}

func (f *condition) Unknown() *condition {
	return f.mutation(func(c *apis.Condition) {
		c.Status = corev1.ConditionUnknown
	})
}

func (f *condition) True() *condition {
	return f.mutation(func(c *apis.Condition) {
		c.Status = corev1.ConditionTrue
		c.Reason = ""
		c.Message = ""
	})
}

func (f *condition) False() *condition {
	return f.mutation(func(c *apis.Condition) {
		c.Status = corev1.ConditionFalse
	})
}

func (f *condition) Reason(reason, message string) *condition {
	return f.mutation(func(c *apis.Condition) {
		c.Reason = reason
		c.Message = message
	})
}

func (f *condition) Info() *condition {
	return f.mutation(func(c *apis.Condition) {
		c.Severity = apis.ConditionSeverityInfo
	})
}

func (f *condition) Warning() *condition {
	return f.mutation(func(c *apis.Condition) {
		c.Severity = apis.ConditionSeverityWarning
	})
}

func (f *condition) Error() *condition {
	return f.mutation(func(c *apis.Condition) {
		c.Severity = apis.ConditionSeverityError
	})
}
