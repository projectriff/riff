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
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/buildpack/pack"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/k8s"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	kailtesting "github.com/projectriff/riff/pkg/testing/kail"
	packtesting "github.com/projectriff/riff/pkg/testing/pack"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFunctionCreateOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldError: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingField(cli.ImageFlagName),
				cli.ErrMissingOneOf(cli.GitRepoFlagName, cli.LocalPathFlagName),
			),
		},
		{
			Name: "git source",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "master",
			},
			ShouldValidate: true,
		},
		{
			Name: "local source",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				LocalPath:       ".",
			},
			ShouldValidate: true,
		},
		{
			Name: "no source",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
			},
			ExpectFieldError: cli.ErrMissingOneOf(cli.GitRepoFlagName, cli.LocalPathFlagName),
		},
		{
			Name: "multiple sources",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "master",
				LocalPath:       ".",
			},
			ExpectFieldError: cli.ErrMultipleOneOf(cli.GitRepoFlagName, cli.LocalPathFlagName),
		},
		{
			Name: "git source with cache",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "master",
				CacheSize:       "8Gi",
			},
			ShouldValidate: true,
		},
		{
			Name: "local source with cache",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				LocalPath:       ".",
				CacheSize:       "8Gi",
			},
			ExpectFieldError: cli.ErrDisallowedFields(cli.CacheSizeFlagName),
		},
		{
			Name: "invalid cache",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "master",
				CacheSize:       "X",
			},
			ExpectFieldError: cli.ErrInvalidValue("X", cli.CacheSizeFlagName),
		},
		{
			Name: "with git subpath",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "master",
				SubPath:         "some/directory",
			},
			ShouldValidate: true,
		},
		{
			Name: "with local subpath",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				LocalPath:       ".",
				SubPath:         "some/directory",
			},
			ExpectFieldError: cli.ErrDisallowedFields(cli.SubPathFlagName),
		},
		{
			Name: "missing git revision",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "",
			},
			ExpectFieldError: cli.ErrMissingField(cli.GitRevisionFlagName),
		},
		{
			Name: "git source, tail",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "master",
				Tail:            true,
				WaitTimeout:     "10m",
			},
			ShouldValidate: true,
		},
		{
			Name: "git source, tail missing timeout",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "master",
				Tail:            true,
			},
			ExpectFieldError: cli.ErrMissingField(cli.WaitTimeoutFlagName),
		},
		{
			Name: "git source, tail invalid timeout",
			Options: &commands.FunctionCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				GitRepo:         "https://example.com/repo.git",
				GitRevision:     "master",
				Tail:            true,
				WaitTimeout:     "d",
			},
			ExpectFieldError: cli.ErrInvalidValue("d", cli.WaitTimeoutFlagName),
		},
	}

	table.Run(t)
}

func TestFunctionCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	functionName := "my-function"
	imageDefault := "_"
	imageTag := "registry.example.com/repo:tag"
	registryHost := "registry.example.com"
	gitRepo := "https://example.com/repo.git"
	gitMaster := "master"
	gitSha := "deadbeefdeadbeefdeadbeefdeadbeef"
	subPath := "some/path"
	cacheSize := "8Gi"
	cacheSizeQuantity := resource.MustParse(cacheSize)
	localPath := "."
	artifact := "test-artifact.js"
	handler := "functions.Handler"
	invoker := "java"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "git repo",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				},
			},
			ExpectOutput: `
Created function "my-function"
`,
		},
		{
			Name: "git repo with revision",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo, cli.GitRevisionFlagName, gitSha},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitSha,
							},
						},
					},
				},
			},
			ExpectOutput: `
Created function "my-function"
`,
		},
		{
			Name: "git repo with subpath",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo, cli.SubPathFlagName, subPath},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
							SubPath: subPath,
						},
					},
				},
			},
			ExpectOutput: `
Created function "my-function"
`,
		},
		{
			Name: "git repo with cache",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo, cli.CacheSizeFlagName, cacheSize},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image:     imageTag,
						CacheSize: &cacheSizeQuantity,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				},
			},
			ExpectOutput: `
Created function "my-function"
`,
		},
		{
			Name: "local path",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.LocalPathFlagName, localPath, cli.ArtifactFlagName, artifact, cli.HandlerFlagName, handler, cli.InvokerFlagName, invoker},
			Prepare: func(t *testing.T, c *cli.Config) error {
				packClient := &packtesting.Client{}
				c.Pack = packClient
				packClient.On("Build", mock.Anything, pack.BuildOptions{
					Image:   imageTag,
					AppDir:  localPath,
					Builder: "projectriff/builder:0.2.0",
					Env: map[string]string{
						"RIFF":          "true",
						"RIFF_ARTIFACT": artifact,
						"RIFF_HANDLER":  handler,
						"RIFF_OVERRIDE": invoker,
					},
					Publish: true,
				}).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...build output...\n")
				})
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				packClient := c.Pack.(*packtesting.Client)
				packClient.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "riff-system",
						Name:      "builders",
					},
					Data: map[string]string{
						"riff-function": "projectriff/builder:0.2.0",
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
						Image:    imageTag,
						Artifact: artifact,
						Handler:  handler,
						Invoker:  invoker,
					},
				},
			},
			ExpectOutput: `
...build output...
Created function "my-function"
`,
		},
		{
			Name: "local path, no builders",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.LocalPathFlagName, localPath},
			ExpectOutput: `
`,
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				if expected, actual := `configmaps "builders" not found`, err.Error(); expected != actual {
					t.Errorf("expected error %q, actual %q", expected, actual)
				}
			},
		},
		{
			Name: "local path, no function builder",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.LocalPathFlagName, localPath},
			GivenObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "riff-system",
						Name:      "builders",
					},
					Data: map[string]string{},
				},
			},
			ExpectOutput: `
`,
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				if expected, actual := `unknown builder for "riff-function"`, err.Error(); expected != actual {
					t.Errorf("expected error %q, actual %q", expected, actual)
				}
			},
		},
		{
			Name: "local path, pack error",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.LocalPathFlagName, localPath, cli.ArtifactFlagName, artifact, cli.HandlerFlagName, handler, cli.InvokerFlagName, invoker},
			Prepare: func(t *testing.T, c *cli.Config) error {
				packClient := &packtesting.Client{}
				c.Pack = packClient
				packClient.On("Build", mock.Anything, pack.BuildOptions{
					Image:   imageTag,
					AppDir:  localPath,
					Builder: "projectriff/builder:0.2.0",
					Env: map[string]string{
						"RIFF":          "true",
						"RIFF_ARTIFACT": artifact,
						"RIFF_HANDLER":  handler,
						"RIFF_OVERRIDE": invoker,
					},
					Publish: true,
				}).Return(fmt.Errorf("pack error")).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...build output...\n")
				})
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				packClient := c.Pack.(*packtesting.Client)
				packClient.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "riff-system",
						Name:      "builders",
					},
					Data: map[string]string{
						"riff-function": "projectriff/builder:0.2.0",
					},
				},
			},
			ExpectOutput: `
...build output...
`,
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				if expected, actual := "pack error", err.Error(); expected != actual {
					t.Errorf("expected error %q, actual %q", expected, actual)
				}
			},
		},
		{
			Name: "local path, default image",
			Args: []string{functionName, cli.LocalPathFlagName, localPath, cli.ArtifactFlagName, artifact, cli.HandlerFlagName, handler, cli.InvokerFlagName, invoker},
			Prepare: func(t *testing.T, c *cli.Config) error {
				packClient := &packtesting.Client{}
				c.Pack = packClient
				packClient.On("Build", mock.Anything, pack.BuildOptions{
					Image:   fmt.Sprintf("%s/%s", registryHost, functionName),
					AppDir:  localPath,
					Builder: "projectriff/builder:0.2.0",
					Env: map[string]string{
						"RIFF":          "true",
						"RIFF_ARTIFACT": artifact,
						"RIFF_HANDLER":  handler,
						"RIFF_OVERRIDE": invoker,
					},
					Publish: true,
				}).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...build output...\n")
				})
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				packClient := c.Pack.(*packtesting.Client)
				packClient.AssertExpectations(t)
				return nil
			},
			GivenObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"default-image-prefix": registryHost,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "riff-system",
						Name:      "builders",
					},
					Data: map[string]string{
						"riff-function": "projectriff/builder:0.2.0",
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
						Image:    imageDefault,
						Artifact: artifact,
						Handler:  handler,
						Invoker:  invoker,
					},
				},
			},
			ExpectOutput: `
...build output...
Created function "my-function"
`,
		},
		{
			Name:        "local path, default image, bad default",
			Args:        []string{functionName, cli.LocalPathFlagName, localPath, cli.ArtifactFlagName, artifact, cli.HandlerFlagName, handler, cli.InvokerFlagName, invoker},
			ShouldError: true,
		},
		{
			Name: "local path, default image",
			Args: []string{functionName, cli.LocalPathFlagName, localPath, cli.ArtifactFlagName, artifact, cli.HandlerFlagName, handler, cli.InvokerFlagName, invoker},
			GivenObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"default-image-prefix": "",
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error existing function",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo},
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
						Source: &buildv1alpha1.Source{
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
			Name: "error during create",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "functions"),
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
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
			Name: "tail logs",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo, cli.TailFlagName},
			Prepare: func(t *testing.T, c *cli.Config) error {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("FunctionLogs", mock.Anything, &buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				}, cli.TailSinceCreateDefault, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...log output...\n")
				})
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				},
			},
			ExpectOutput: `
Created function "my-function"
...log output...
`,
		},
		{
			Name: "tail timeout",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo, cli.TailFlagName, cli.WaitTimeoutFlagName, "1ms"},
			Prepare: func(t *testing.T, c *cli.Config) error {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("FunctionLogs", mock.Anything, &buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				}, cli.TailSinceCreateDefault, mock.Anything).Return(k8s.ErrWaitTimeout).Run(func(args mock.Arguments) {
					ctx := args[0].(context.Context)
					fmt.Fprintf(c.Stdout, "...log output...\n")
					// wait for context to be cancelled, plus some fudge
					<-ctx.Done()
					time.Sleep(time.Millisecond)
				})
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				},
			},
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				if expected, actual := k8s.ErrWaitTimeout, err; expected != actual {
					t.Errorf("expected error %q, actual %q", expected, actual)
				}
				for _, line := range []string{
					`
Created function "my-function"
`,
					`
...log output...
`,
					`
Timeout after "1ms" waiting for "my-function" to become ready
To view status run: riff function list --namespace default
To continue watching logs run: riff function tail my-function --namespace default
`,
				} {
					if expected, actual := line[1:], output; !strings.Contains(actual, expected) {
						t.Errorf("expected output to contain %q, actual %q", expected, actual)
					}
				}
			},
		},
		{
			Name: "tail error",
			Args: []string{functionName, cli.ImageFlagName, imageTag, cli.GitRepoFlagName, gitRepo, cli.TailFlagName},
			Prepare: func(t *testing.T, c *cli.Config) error {
				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("FunctionLogs", mock.Anything, &buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
				}, cli.TailSinceCreateDefault, mock.Anything).Return(fmt.Errorf("kail error"))
				return nil
			},
			CleanUp: func(t *testing.T, c *cli.Config) error {
				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
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
