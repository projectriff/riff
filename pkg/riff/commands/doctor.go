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

package commands

import (
	"context"
	"fmt"

	"github.com/knative/pkg/apis"
	"github.com/projectriff/riff/pkg/riff"
	"github.com/spf13/cobra"
)

type DoctorOptions struct {
	Namespace string
}

func (opt *DoctorOptions) Validate(ctx context.Context) *apis.FieldError {
	// TODO implement
	return nil
}

func NewDoctorCommand(c *riff.Config) *cobra.Command {
	opt := &DoctorOptions{}

	cmd := &cobra.Command{
		Use:     "doctor",
		Args:    riff.Args(),
		PreRunE: riff.ValidateOptions(opt),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}

	riff.NamespaceFlag(cmd, c, &opt.Namespace)

	return cmd
}
