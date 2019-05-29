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

package cli

import (
	"time"

	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

func FormatTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return Swarnf("<unknown>")
	}
	return duration.HumanDuration(time.Since(timestamp.Time))
}

func FormatEmptyString(str string) string {
	if str == "" {
		return Sfaintf("<empty>")
	}
	return str
}

func FormatConditionStatus(cond *duckv1alpha1.Condition) string {
	if cond == nil || cond.Status == "" {
		return Swarnf("<unknown>")
	}
	status := string(cond.Status)
	switch status {
	case "True":
		return Ssuccessf(string(cond.Type))
	case "False":
		if cond.Reason == "" {
			// display something if there is no reason
			return Serrorf("not-" + string(cond.Type))
		}
		return Serrorf(cond.Reason)
	default:
		return Sinfof(status)
	}
}

func FormatConditionMessage(cond *duckv1alpha1.Condition) string {
	switch {
	case cond == nil:
		return Swarnf("<unknown>")
	case cond.Message == "":
		return FormatEmptyString(cond.Message)
	case cond.IsFalse():
		return Serrorf(cond.Message)
	case cond.IsTrue():
		return Ssuccessf(cond.Message)
	case cond.IsUnknown():
		return Sinfof(cond.Message)
	default:
		return Swarnf(cond.Message)
	}
}
