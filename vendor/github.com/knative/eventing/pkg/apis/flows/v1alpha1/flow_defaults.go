/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

func (f *Flow) SetDefaults() {
	f.Spec.SetDefaults()
}

func (fs *FlowSpec) SetDefaults() {
	fs.Action.SetDefaults()
	fs.Trigger.SetDefaults()
}

func (fa *FlowAction) SetDefaults() {
	// TODO anything?
}

func (et *EventTrigger) SetDefaults() {
	// TODO anything?
}
