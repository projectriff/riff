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

package options

import (
	"context"

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/validation"
)

type ListOptions struct {
	Namespace     string
	AllNamespaces bool
}

func (opts *ListOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	if opts.Namespace == "" && !opts.AllNamespaces {
		errs = errs.Also(cli.ErrMissingOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName))
	}
	if opts.Namespace != "" && opts.AllNamespaces {
		errs = errs.Also(cli.ErrMultipleOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName))
	}

	return errs
}

type ResourceOptions struct {
	Namespace string
	Name      string
}

func (opts *ResourceOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	if opts.Namespace == "" {
		errs = errs.Also(cli.ErrMissingField(cli.NamespaceFlagName))
	}

	if opts.Name == "" {
		errs = errs.Also(cli.ErrMissingField(cli.NameArgumentName))
	} else {
		errs = errs.Also(validation.K8sName(opts.Name, cli.NameArgumentName))
	}

	return errs
}

type DeleteOptions struct {
	Namespace string
	Names     []string
	All       bool
}

func (opts *DeleteOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	if opts.Namespace == "" {
		errs = errs.Also(cli.ErrMissingField(cli.NamespaceFlagName))
	}

	if opts.All && len(opts.Names) != 0 {
		errs = errs.Also(cli.ErrMultipleOneOf(cli.AllFlagName, cli.NamesArgumentName))
	}
	if !opts.All && len(opts.Names) == 0 {
		errs = errs.Also(cli.ErrMissingOneOf(cli.AllFlagName, cli.NamesArgumentName))
	}

	errs = errs.Also(validation.K8sNames(opts.Names, cli.NamesArgumentName))

	return errs
}
