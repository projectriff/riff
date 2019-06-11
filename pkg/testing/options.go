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
	"reflect"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/pkg/cli"
)

const TestField = "test-field"

var (
	ValidListOptions = cli.ListOptions{
		Namespace: "default",
	}
	InvalidListOptions           = cli.ListOptions{}
	InvalidListOptionsFieldError = cli.ErrMissingOneOf(cli.NamespaceFlagName, cli.AllNamespacesFlagName)
)

var (
	ValidResourceOptions = cli.ResourceOptions{
		Namespace: "default",
		Name:      "push-credentials",
	}
	InvalidResourceOptions           = cli.ResourceOptions{}
	InvalidResourceOptionsFieldError = cli.EmptyFieldError.Also(
		cli.ErrMissingField(cli.NamespaceFlagName),
		cli.ErrMissingField(cli.NameArgumentName),
	)
)

var (
	ValidDeleteOptions = cli.DeleteOptions{
		Namespace: "default",
		Names:     []string{"my-resource"},
	}
	InvalidDeleteOptions = cli.DeleteOptions{
		Namespace: "default",
	}
	InvalidDeleteOptionsFieldError = cli.ErrMissingOneOf(cli.AllFlagName, cli.NamesArgumentName)
)

func DiffFieldErrors(expected, actual *cli.FieldError) string {
	return cmp.Diff(flattenFieldErrors(expected), flattenFieldErrors(actual), compareFieldError)
}

var compareFieldError = cmp.Comparer(func(a, b cli.FieldError) bool {
	if a.Message != b.Message {
		return false
	}
	if a.Details != b.Details {
		return false
	}
	return cmp.Equal(filterEmpty(a.Paths), filterEmpty(b.Paths))
})

func filterEmpty(s []string) []string {
	r := []string{}
	for _, i := range s {
		if i != "" {
			r = append(r, i)
		}
	}
	return r
}

func flattenFieldErrors(err *cli.FieldError) []cli.FieldError {
	errs := []cli.FieldError{}

	if err == nil {
		return errs
	}

	if err.Message != "" {
		errs = append(errs, *err)
	}
	for _, nestedErr := range extractNestedErrors(err) {
		errs = append(errs, flattenFieldErrors(&nestedErr)...)
	}

	return errs
}

func extractNestedErrors(err *cli.FieldError) []cli.FieldError {
	var nestedErrors []cli.FieldError

	// `nestedErrors = err.errors`
	// TODO let's get this exposed on the type so we don't need to do unsafe reflection
	ev := reflect.ValueOf(err).Elem().FieldByName("errors")
	ev = reflect.NewAt(ev.Type(), unsafe.Pointer(ev.UnsafeAddr())).Elem()
	nev := reflect.ValueOf(&nestedErrors).Elem()
	nev = reflect.NewAt(nev.Type(), unsafe.Pointer(nev.UnsafeAddr())).Elem()
	nev.Set(ev)

	return nestedErrors
}
