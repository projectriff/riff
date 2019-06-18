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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/system/pkg/apis/build"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialApplyOptions struct {
	cli.ResourceOptions

	DockerHubId       string
	DockerHubPassword []byte

	GcrTokenPath string

	Registry         string
	RegistryUser     string
	RegistryPassword []byte

	DefaultImagePrefix    string
	SetDefaultImagePrefix bool

	DryRun bool
}

var (
	_ cli.Validatable = (*CredentialApplyOptions)(nil)
	_ cli.Executable  = (*CredentialApplyOptions)(nil)
	_ cli.DryRunable  = (*CredentialApplyOptions)(nil)
)

func (opts *CredentialApplyOptions) Validate(ctx context.Context) *cli.FieldError {
	errs := cli.EmptyFieldError

	errs = errs.Also(opts.ResourceOptions.Validate(ctx))

	// docker-hub, gcr and registry are mutually exclusive
	used := []string{}
	unused := []string{}

	if opts.DockerHubId != "" {
		used = append(used, cli.DockerHubFlagName)
	} else {
		unused = append(unused, cli.DockerHubFlagName)
	}

	if opts.GcrTokenPath != "" {
		used = append(used, cli.GcrFlagName)
	} else {
		unused = append(unused, cli.GcrFlagName)
	}

	if opts.Registry != "" {
		used = append(used, cli.RegistryFlagName)
	} else {
		unused = append(unused, cli.RegistryFlagName)
	}

	if len(used) == 0 {
		errs = errs.Also(cli.ErrMissingOneOf(unused...))
	} else if len(used) > 1 {
		errs = errs.Also(cli.ErrMultipleOneOf(used...))
	}

	if opts.DockerHubId != "" && len(opts.DockerHubPassword) == 0 {
		errs = errs.Also(cli.ErrMissingField("<docker-hub-password>"))
	}

	if len(opts.RegistryPassword) != 0 && opts.RegistryUser == "" {
		errs = errs.Also(cli.ErrMissingField(cli.RegistryUserFlagName))
	}

	if opts.SetDefaultImagePrefix && opts.DefaultImagePrefix == "" && opts.Registry != "" {
		errs = errs.Also(cli.ErrInvalidValue(fmt.Sprintf("cannot be used with %s, without %s", cli.RegistryFlagName, cli.DefaultImagePrefixFlagName), cli.SetDefaultImagePrefixFlagName))
	}

	return errs
}

func (opts *CredentialApplyOptions) Exec(ctx context.Context, c *cli.Config) error {
	// get desired credential and image prefix
	secret, imagePrefix, err := makeCredential(opts)
	if err != nil {
		return err
	}

	if err := applyCredential(ctx, c, opts, secret); err != nil {
		return err
	}
	c.Successf("Apply credentials %q\n", opts.Name)

	if opts.DefaultImagePrefix != "" || opts.SetDefaultImagePrefix {
		if opts.DefaultImagePrefix != "" {
			imagePrefix = opts.DefaultImagePrefix
		}
		if imagePrefix == "" {
			// guarded by opts.Validate()
			c.Infof("Unable to derive default image prefix\n")
		} else {
			err := setDefaultImagePrefix(ctx, c, opts, imagePrefix)
			if err != nil {
				return err
			}
			c.Successf("Set default image prefix to %q\n", imagePrefix)
		}
	}

	return nil
}

func (opts *CredentialApplyOptions) IsDryRun() bool {
	return opts.DryRun
}

func NewCredentialApplyCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &CredentialApplyOptions{}

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "create or update credentials for a container registry",
		Long: strings.TrimSpace(`
<todo>
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s credential apply my-docker-hub-creds %s my-docker-id", c.Name, cli.DockerHubFlagName),
			fmt.Sprintf("%s credential apply my-docker-hub-creds %s my-docker-id %s", c.Name, cli.DockerHubFlagName, cli.SetDefaultImagePrefixFlagName),
			fmt.Sprintf("%s credential apply my-gcr-creds %s path/to/token.json", c.Name, cli.GcrFlagName),
			fmt.Sprintf("%s credential apply my-gcr-creds %s path/to/token.json %s", c.Name, cli.GcrFlagName, cli.SetDefaultImagePrefixFlagName),
			fmt.Sprintf("%s credential apply my-registry-creds %s http://registry.example.com %s my-username", c.Name, cli.RegistryFlagName, cli.RegistryUserFlagName),
			fmt.Sprintf("%s credential apply my-registry-creds %s http://registry.example.com %s my-username %s registry.example.com/my-username", c.Name, cli.RegistryFlagName, cli.RegistryUserFlagName, cli.DefaultImagePrefixFlagName),
		}, "\n"),
		Args: cli.Args(
			cli.NameArg(&opts.Name),
		),
		PreRunE: cli.Sequence(
			func(cmd *cobra.Command, args []string) error {
				if opts.DockerHubId != "" {
					return cli.ReadStdin(c, &opts.DockerHubPassword, "Docker Hub password")(cmd, args)
				}
				if opts.RegistryUser != "" {
					return cli.ReadStdin(c, &opts.RegistryPassword, "Registry password")(cmd, args)
				}
				return nil
			},
			cli.ValidateOptions(ctx, opts),
		),
		RunE: cli.ExecOptions(ctx, c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)
	cmd.Flags().StringVar(&opts.DockerHubId, cli.StripDash(cli.DockerHubFlagName), "", "Docker Hub `username`, the password must be provided via stdin")
	cmd.Flags().StringVar(&opts.GcrTokenPath, cli.StripDash(cli.GcrFlagName), "", "path to Google Container Registry service account token `file`")
	cmd.Flags().StringVar(&opts.Registry, cli.StripDash(cli.RegistryFlagName), "", "registry `url`")
	cmd.Flags().StringVar(&opts.RegistryUser, cli.StripDash(cli.RegistryUserFlagName), "", "`username` for a registry, the password must be provided via stdin")
	cmd.Flags().StringVar(&opts.DefaultImagePrefix, cli.StripDash(cli.DefaultImagePrefixFlagName), "", fmt.Sprintf("default `repository` prefix for built images, implies %s", cli.SetDefaultImagePrefixFlagName))
	cmd.Flags().BoolVar(&opts.SetDefaultImagePrefix, cli.StripDash(cli.SetDefaultImagePrefixFlagName), false, "use this registry as the default for built images")
	cmd.Flags().BoolVar(&opts.DryRun, cli.StripDash(cli.DryRunFlagName), false, "print kubernetes resources to stdout rather than apply them to the cluster, messages normally on stdout will be sent to stderr")

	return cmd
}

func makeCredential(opts *CredentialApplyOptions) (*corev1.Secret, string, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: opts.Namespace,
			Name:      opts.Name,
		},
	}
	defaultPrefix := ""

	switch {
	case opts.DockerHubId != "":
		secret.Labels = map[string]string{
			build.CredentialLabelKey: "docker-hub",
		}
		secret.Annotations = map[string]string{
			"build.knative.dev/docker-0": "https://index.docker.io/v1/",
		}
		secret.Type = corev1.SecretTypeBasicAuth
		secret.StringData = map[string]string{
			"username": opts.DockerHubId,
			"password": string(opts.DockerHubPassword),
		}
		defaultPrefix = fmt.Sprintf("docker.io/%s", opts.DockerHubId)

	case opts.GcrTokenPath != "":
		token, err := ioutil.ReadFile(opts.GcrTokenPath)
		if err != nil {
			return nil, "", err
		}
		secret.Labels = map[string]string{
			build.CredentialLabelKey: "gcr",
		}
		secret.Annotations = map[string]string{
			"build.knative.dev/docker-0": "https://gcr.io",
			"build.knative.dev/docker-1": "https://us.gcr.io",
			"build.knative.dev/docker-2": "https://eu.gcr.io",
			"build.knative.dev/docker-3": "https://asia.gcr.io",
		}
		secret.Type = corev1.SecretTypeBasicAuth
		secret.StringData = map[string]string{
			"username": "_json_key",
			"password": string(token),
		}
		tokenMap := map[string]string{}
		err = json.Unmarshal(token, &tokenMap)
		if err != nil {
			return nil, "", err
		}
		defaultPrefix = fmt.Sprintf("gcr.io/%s", tokenMap["project_id"])

	case opts.RegistryUser != "":
		secret.Labels = map[string]string{
			build.CredentialLabelKey: "basic-auth",
		}
		secret.Annotations = map[string]string{
			"build.knative.dev/docker-0": opts.Registry,
		}
		secret.Type = corev1.SecretTypeBasicAuth
		secret.StringData = map[string]string{
			"username": opts.RegistryUser,
			"password": string(opts.RegistryPassword),
		}
		// unable to determine default prefix for registry
	}

	return secret, defaultPrefix, nil
}

func setDefaultImagePrefix(ctx context.Context, c *cli.Config, opts *CredentialApplyOptions, defaultImagePrefix string) error {
	configMapName := "riff-build"
	defaultImagePrefixKey := "default-image-prefix"

	riffBuildConfig, err := c.Core().ConfigMaps(opts.Namespace).Get(configMapName, metav1.GetOptions{})
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}

		// create riff-build configmaps
		riffBuildConfig := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: opts.Namespace,
				Name:      configMapName,
			},
			Data: map[string]string{
				defaultImagePrefixKey: defaultImagePrefix,
			},
		}
		if opts.DryRun {
			cli.DryRunResource(ctx, riffBuildConfig, corev1.SchemeGroupVersion.WithKind("ConfigMap"))
		} else {
			_, err := c.Core().ConfigMaps(opts.Namespace).Create(riffBuildConfig)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// update riff-build config
	riffBuildConfig = riffBuildConfig.DeepCopy()
	riffBuildConfig.Data[defaultImagePrefixKey] = defaultImagePrefix
	if opts.DryRun {
		cli.DryRunResource(ctx, riffBuildConfig, corev1.SchemeGroupVersion.WithKind("ConfigMap"))
	} else {
		_, err := c.Core().ConfigMaps(opts.Namespace).Update(riffBuildConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyCredential(ctx context.Context, c *cli.Config, opts *CredentialApplyOptions, desiredSecret *corev1.Secret) error {
	// look for existing secret
	existing, err := c.Core().Secrets(opts.Namespace).Get(opts.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return err
		}

		// create secret
		if opts.DryRun {
			cli.DryRunResource(ctx, desiredSecret, corev1.SchemeGroupVersion.WithKind("Secret"))
		} else {
			_, err := c.Core().Secrets(opts.Namespace).Create(desiredSecret)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// ensure we are not mutating a non-riff secret
	if _, ok := existing.Labels[build.CredentialLabelKey]; !ok {
		return fmt.Errorf("credential %q exists, but is not owned by riff", opts.Name)
	}

	// update existing secret
	existing = existing.DeepCopy()
	existing.Labels[build.CredentialLabelKey] = desiredSecret.Labels[build.CredentialLabelKey]
	existing.Annotations = desiredSecret.Annotations
	existing.Type = desiredSecret.Type
	existing.StringData = desiredSecret.StringData
	existing.Data = desiredSecret.Data
	if opts.DryRun {
		cli.DryRunResource(ctx, existing, corev1.SchemeGroupVersion.WithKind("Secret"))
	} else {
		_, err := c.Core().Secrets(opts.Namespace).Update(existing)
		if err != nil {
			return err
		}
	}

	return nil
}
