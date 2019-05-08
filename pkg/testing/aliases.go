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
	"testing"

	kntesting "github.com/knative/pkg/reconciler/testing"
	clientgotesting "k8s.io/client-go/testing"
)

type T = testing.T

type ObjectSorter = kntesting.ObjectSorter

var (
	NewObjectSorter = kntesting.NewObjectSorter
	InduceFailure   = kntesting.InduceFailure
	ValidateCreates = kntesting.ValidateCreates
	ValidateUpdates = kntesting.ValidateUpdates
)

type Action = clientgotesting.Action
type ActionRecorder = kntesting.ActionRecorder
type ActionRecorderList = kntesting.ActionRecorderList

type ReactionFunc = clientgotesting.ReactionFunc
type ActionImpl = clientgotesting.ActionImpl
type DeleteAction = clientgotesting.DeleteAction
type DeleteCollectionAction = clientgotesting.DeleteCollectionAction
