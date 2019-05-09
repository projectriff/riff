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

	"github.com/knative/pkg/apis"
	"github.com/projectriff/riff/pkg/riff"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialDeleteOptions struct {
	Namespace string
	Names     []string
	All       bool
}

func (opt *CredentialDeleteOptions) Validate(ctx context.Context) *apis.FieldError {
	errs := &apis.FieldError{}

	if opt.Namespace == "" {
		errs = errs.Also(apis.ErrMissingField("namespace"))
	}

	if opt.All && len(opt.Names) != 0 {
		errs = errs.Also(apis.ErrMultipleOneOf("all", "names"))
	}
	if !opt.All && len(opt.Names) == 0 {
		errs = errs.Also(apis.ErrMissingOneOf("all", "names"))
	}

	for i, name := range opt.Names {
		if out := validation.NameIsDNSSubdomain(name, false); len(out) != 0 || name == "" {
			// TODO capture info about why the name is invalid
			errs = errs.Also(apis.ErrInvalidArrayValue(name, "names", i))
		}
	}

	return errs
}

func NewCredentialDeleteCommand(c *riff.Config) *cobra.Command {
	opt := &CredentialDeleteOptions{}

	cmd := &cobra.Command{
		Use: "delete",
		Args: riff.Args(
			riff.NamesArg(&opt.Names),
		),
		PreRunE: riff.ValidateOptions(opt),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := c.Core().Secrets(opt.Namespace)

			if opt.All {
				return client.DeleteCollection(nil, metav1.ListOptions{
					// TODO get label from riff system
					LabelSelector: "projectriff.io/credential",
				})
			}

			for _, name := range opt.Names {
				// TODO check for the matching label before deleting
				if err := client.Delete(name, nil); err != nil {
					return err
				}
			}

			return nil
		},
	}

	riff.NamespaceFlag(cmd, c, &opt.Namespace)
	cmd.Flags().BoolVar(&opt.All, "all", false, "delete all secrets in the namespace")

	return cmd
}
