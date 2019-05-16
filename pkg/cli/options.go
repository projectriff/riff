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
	"context"

	"github.com/knative/pkg/apis"
	"github.com/projectriff/riff/pkg/validation"
	"github.com/spf13/cobra"
)

// ValidateOptions bridges a cobra RunE function to the Validatable interface.  All flags and
// arguments must already be bound, with explicit or default values, to the options struct being
// validated. This function is typically used to define the PreRunE phase of a command.
//
// ```
// cmd := &cobra.Command{
// 	   ...
// 	   PreRunE: cli.ValidateOptions(opts),
// }
// ```
func ValidateOptions(opts apis.Validatable) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		if err := opts.Validate(ctx); err.Error() != "" {
			return err
		}
		cmd.SilenceUsage = true
		return nil
	}
}

type ListOptions struct {
	Namespace     string
	AllNamespaces bool
}

func (opts *ListOptions) Validate(ctx context.Context) *FieldError {
	errs := &FieldError{}

	if opts.Namespace == "" && !opts.AllNamespaces {
		errs = errs.Also(ErrMissingOneOf(NamespaceFlagName, AllNamespacesFlagName))
	}
	if opts.Namespace != "" && opts.AllNamespaces {
		errs = errs.Also(ErrMultipleOneOf(NamespaceFlagName, AllNamespacesFlagName))
	}

	return errs
}

type ResourceOptions struct {
	Namespace string
	Name      string
}

func (opts *ResourceOptions) Validate(ctx context.Context) *FieldError {
	errs := &FieldError{}

	if opts.Namespace == "" {
		errs = errs.Also(ErrMissingField(NamespaceFlagName))
	}

	if opts.Name == "" {
		errs = errs.Also(ErrMissingField(opts.Name, NameArgumentName))
	} else {
		errs = errs.Also(validation.K8sName(opts.Name, NameArgumentName))
	}

	return errs
}

type DeleteOptions struct {
	Namespace string
	Names     []string
	All       bool
}

func (opts *DeleteOptions) Validate(ctx context.Context) *FieldError {
	errs := &FieldError{}

	if opts.Namespace == "" {
		errs = errs.Also(ErrMissingField(NamespaceFlagName))
	}

	if opts.All && len(opts.Names) != 0 {
		errs = errs.Also(ErrMultipleOneOf(AllFlagName, NamesArgumentName))
	}
	if !opts.All && len(opts.Names) == 0 {
		errs = errs.Also(ErrMissingOneOf(AllFlagName, NamesArgumentName))
	}

	errs = errs.Also(validation.K8sNames(opts.Names, NamesArgumentName))

	return errs
}
