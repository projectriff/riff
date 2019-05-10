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

package commands_test

import (
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFunctionCreateOptions(t *testing.T) {
	defaultOptions := func() cli.Validatable {
		return &commands.FunctionCreateOptions{
			Namespace:   "default",
			GitRevision: "master",
		}
	}

	table := testing.OptionsTable{
		{
			Name: "default",
			ExpectErrors: []cli.FieldError{
				*cli.ErrMissingField("name"),
				*cli.ErrMissingField("image"),
				*cli.ErrMissingOneOf("git-repo", "local-path"),
			},
		},
		{
			Name: "git source",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.GitRepo = "https://example.com/repo.git"
			},
			ShouldValidate: true,
		},
		{
			Name: "local source",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.LocalPath = "."
			},
			ShouldValidate: true,
		},
		{
			Name: "no source",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
			},
			ExpectErrors: []cli.FieldError{
				*cli.ErrMissingOneOf("git-repo", "local-path"),
			},
		},
		{
			Name: "multiple sources",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.GitRepo = "https://example.com/repo.git"
				opts.LocalPath = "."
			},
			ExpectErrors: []cli.FieldError{
				*cli.ErrMultipleOneOf("git-repo", "local-path"),
			},
		},
		{
			Name: "git source with cache",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.GitRepo = "https://example.com/repo.git"
				opts.CacheSize = "8Gi"
			},
			ShouldValidate: true,
		},
		{
			Name: "local source with cache",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.LocalPath = "."
				opts.CacheSize = "8Gi"
			},
			ExpectErrors: []cli.FieldError{
				*cli.ErrDisallowedFields("cache-size"),
			},
		},
		{
			Name: "invalid cache",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.GitRepo = "https://example.com/repo.git"
				opts.CacheSize = "X"
			},
			ExpectErrors: []cli.FieldError{
				*cli.ErrInvalidValue("X", "cache-size"),
			},
		},
		{
			Name: "with git subpath",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.GitRepo = "https://example.com/repo.git"
				opts.SubPath = "some/directory"
			},
			ShouldValidate: true,
		},
		{
			Name: "with local subpath",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.LocalPath = "."
				opts.SubPath = "some/directory"
			},
			ExpectErrors: []cli.FieldError{
				*cli.ErrDisallowedFields("sub-path"),
			},
		},
		{
			Name: "missing namespace",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Namespace = ""
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.GitRepo = "https://example.com/repo.git"
			},
			ExpectErrors: []cli.FieldError{
				*cli.ErrMissingField("namespace"),
			},
		},
		{
			Name: "missing git revision",
			OverrideOptions: func(opts *commands.FunctionCreateOptions) {
				opts.Name = "my-function"
				opts.Image = "example.com/repo:tag"
				opts.GitRepo = "https://example.com/repo.git"
				opts.GitRevision = ""
			},
			ExpectErrors: []cli.FieldError{
				*cli.ErrMissingField("git-revision"),
			},
		},
	}

	table.Run(t, defaultOptions)
}

func TestFunctionCreateCommand(t *testing.T) {
	t.Parallel()

	defaultNamespace := "default"
	functionName := "my-function"
	imageTag := "registry.example.com/repo:tag"
	imageDigest := "registry.example.com/repo@sha256:deadbeefdeadbeefdeadbeefdeadbeef"
	gitRepo := "https://example.com/repo.git"
	gitMaster := "master"
	gitSha := "deadbeefdeadbeefdeadbeefdeadbeef"
	subPath := "some/path"
	cacheSize := "8Gi"
	cacheSizeQuantity := resource.MustParse(cacheSize)
	localPath := "."

	table := testing.CommandTable{
		{
			Name:        "empty",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "git repo",
			Args: []string{functionName, "--image", imageTag, "--git-repo", gitRepo},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				},
			},
		},
		{
			Name: "git repo with revision",
			Args: []string{functionName, "--image", imageTag, "--git-repo", gitRepo, "--git-revision", gitSha},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitSha,
							},
						},
					},
				},
			},
		},
		{
			Name: "git repo with subpath",
			Args: []string{functionName, "--image", imageTag, "--git-repo", gitRepo, "--sub-path", subPath},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
							SubPath: subPath,
						},
					},
				},
			},
		},
		{
			Name: "git repo with cache",
			Args: []string{functionName, "--image", imageTag, "--git-repo", gitRepo, "--cache-size", cacheSize},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image:     imageTag,
						CacheSize: &cacheSizeQuantity,
						Source: buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				},
			},
		},
		{
			// TODO impelement
			Skip: true,
			Name: "local path",
			Args: []string{functionName, "--image", imageTag, "--local-path", localPath},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
					},
					Status: buildv1alpha1.FunctionStatus{
						BuildStatus: buildv1alpha1.BuildStatus{
							LatestImage: imageDigest,
						},
					},
				},
			},
		},
		{
			Name: "error existing function",
			Args: []string{functionName, "--image", imageTag, "--git-repo", gitRepo},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error durring create",
			Args: []string{functionName, "--image", imageTag, "--git-repo", gitRepo},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("create", "functions"),
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				},
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewFunctionCreateCommand)
}
