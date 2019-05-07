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
	"bytes"
	"testing"

	"github.com/projectriff/riff/pkg/riff"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

type Table []TableRow

type TableRow struct {
	Name         string
	Args         []string
	Params       *riff.Params
	Objects      []runtime.Object
	WithReactors []clientgotesting.ReactionFunc
	// WantCreates           []metav1.Object
	// WantUpdates           []clientgotesting.UpdateActionImpl
	// WantDeletes           []clientgotesting.DeleteActionImpl
	// WantDeleteCollections []clientgotesting.DeleteCollectionActionImpl
	WantError  bool
	WithOutput func(*testing.T, string, error)
}

func (tests Table) Run(t *testing.T, cmdFactory func(*riff.Params) *cobra.Command) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			p := test.Params
			if p == nil {
				p = &riff.Params{}
			}
			client := NewClient(test.Objects...)
			p.Client = client

			for i := range test.WithReactors {
				// in reverse order since we prepend
				reactor := test.WithReactors[len(test.WithReactors)-1-i]
				client.PrependReactor("*", "*", reactor)
			}

			cmd := cmdFactory(p)
			output := &bytes.Buffer{}

			cmd.SetArgs(test.Args)
			cmd.SetOutput(output)

			err := cmd.Execute()

			if want, got := test.WantError, err != nil; want != got {
				if want {
					t.Errorf("expected command to error, got %v", got)
				} else {
					t.Errorf("expected command not to error, got %v", got)
				}
			}
			if test.WithOutput != nil {
				test.WithOutput(t, output.String(), err)
			}
		})
	}
}
