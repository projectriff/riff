/*
 * Copyright 2018-2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"fmt"
	"github.com/pivotal/go-ape/pkg/furl"
	"github.com/projectriff/riff/pkg/core/tasks"
	"github.com/projectriff/riff/pkg/env"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"sort"
	"strings"
)

const BuildServiceAccountName = "riff-build"

type secretType int

const (
	secretTypeNone secretType = iota
	secretTypeUserProvided
	secretTypeGcr
	secretTypeDockerHub
	secretTypeBasicAuth
)

type NamespaceInitOptions struct {
	NamespaceName string
	Manifest      string

	ImagePrefix string

	NoSecret     bool
	SecretName   string
	GcrTokenPath string
	DockerHubId  string

	Registry     string
	RegistryUser string
}

type NamespaceCleanupOptions struct {
	NamespaceName   string
	SecretName      string
	RemoveNamespace bool
}

func (o *NamespaceInitOptions) initSecretType() secretType {
	switch {
	case o.NoSecret:
		return secretTypeNone
	case o.DockerHubId != "":
		return secretTypeDockerHub
	case o.GcrTokenPath != "":
		return secretTypeGcr
	case o.RegistryUser != "":
		return secretTypeBasicAuth
	default:
		return secretTypeUserProvided
	}
}

func (c *client) explicitOrConfigNamespace(explicitNamespace string) string {
	if explicitNamespace != "" {
		return explicitNamespace
	}

	namespace, _, _ := c.clientConfig.Namespace() // TODO: handle any error
	return namespace
}

func (c *client) NamespaceInit(manifests map[string]*Manifest, options NamespaceInitOptions) error {
	manifest, err := ResolveManifest(manifests, options.Manifest)
	if err != nil {
		return err
	}
	if _, err = c.initNamespace(options.NamespaceName); err != nil {
		return err
	}
	initLabels := getInitLabels()
	if err = c.createInitialSecret(&options, initLabels); err != nil {
		return err
	}
	if err = c.initServiceAccount(&options, initLabels); err != nil {
		return err
	}
	if err = c.initImagePrefix(&options); err != nil {
		return err
	}
	return c.applyManifest(manifest, &options, initLabels)
}

func (c *client) NamespaceCleanup(options NamespaceCleanupOptions) error {
	ns := options.NamespaceName
	initLabelKeys := sortedKeysOf(getInitLabels())
	initLabelSelector := existsSelectors(initLabelKeys)

	fmt.Printf("Deleting serviceaccounts matching label keys %v in namespace %q\n", initLabelKeys, ns)
	if err := c.deleteMatchingServiceAccounts(ns, initLabelSelector); err != nil {
		return err
	}

	fmt.Printf("Deleting secrets matching label keys %v in namespace %q\n", initLabelKeys, ns)
	if err := c.deleteMatchingSecrets(ns, initLabelSelector); err != nil {
		return err
	}

	fmt.Printf("Deleting configmaps matching label keys %v in namespace %q\n", initLabelKeys, ns)
	if err := c.deleteMatchingConfigMaps(ns, initLabelSelector); err != nil {
		return err
	}

	if options.RemoveNamespace {
		fmt.Printf("Deleting namespace %q\n", ns)
		if err := c.kubeClient.CoreV1().Namespaces().Delete(ns, &v1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) initNamespace(ns string) (*corev1.Namespace, error) {
	fmt.Printf("Initializing namespace %q\n\n", ns)
	namespace, err := c.kubeClient.CoreV1().Namespaces().Get(ns, v1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Creating namespace %q \n", ns)
		namespace = &corev1.Namespace{}
		namespace.Name = ns
		namespace, err = c.kubeClient.CoreV1().Namespaces().Create(namespace)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return namespace, nil
}

func (c *client) createInitialSecret(options *NamespaceInitOptions, initLabels map[string]string) error {
	secret, err := c.convertInitialSecret(options, initLabels)
	if err != nil {
		return err
	}
	if secret == nil {
		return nil
	}
	return c.resetBasicSecret(options.NamespaceName, secret)
}

func (c *client) convertInitialSecret(options *NamespaceInitOptions, initLabels map[string]string) (*corev1.Secret, error) {
	namespace := options.NamespaceName
	secretName := options.SecretName
	switch options.initSecretType() {
	case secretTypeGcr:
		return c.convertGcrSecret(namespace, secretName, options.GcrTokenPath, initLabels)
	case secretTypeDockerHub:
		return c.convertDockerHubSecret(namespace, secretName, options.DockerHubId, initLabels)
	case secretTypeBasicAuth:
		return c.convertRegistrySecret(namespace, secretName, options.RegistryUser, options.Registry, initLabels)
	case secretTypeUserProvided:
		err := c.checkSecretExists(options)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil, nil
}

func (c *client) initServiceAccount(options *NamespaceInitOptions, initLabels map[string]string) error {
	ns := options.NamespaceName
	serviceAccount, err := c.kubeClient.CoreV1().ServiceAccounts(ns).Get(BuildServiceAccountName, v1.GetOptions{})
	if errors.IsNotFound(err) {
		serviceAccount = &corev1.ServiceAccount{}
		serviceAccount.Name = BuildServiceAccountName
		serviceAccount.Labels = initLabels
		if options.initSecretType() != secretTypeNone {
			secretName := options.SecretName
			serviceAccount.Secrets = append(serviceAccount.Secrets, corev1.ObjectReference{Name: secretName})
			fmt.Printf("Creating serviceaccount %q using secret %q in namespace %q\n", serviceAccount.Name, secretName, ns)
		} else {
			fmt.Printf("Creating unauthenticated serviceaccount %q in namespace %q\n", serviceAccount.Name, ns)
		}
		if _, err = c.kubeClient.CoreV1().ServiceAccounts(ns).Create(serviceAccount); err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	if options.initSecretType() != secretTypeNone {
		secretName := options.SecretName
		return c.appendNewSecretToServiceAccount(ns, secretName, serviceAccount)
	}
	return nil
}

func (c *client) initImagePrefix(options *NamespaceInitOptions) error {
	imagePrefix, err := DetermineImagePrefix(options.ImagePrefix, options.DockerHubId, options.GcrTokenPath)
	if err != nil {
		return err
	}

	if imagePrefix != "" {
		return c.SetDefaultBuildImagePrefix(options.NamespaceName, imagePrefix)
	}

	fmt.Printf("No image prefix set, resetting possibly existing ones. The --image argument will be required for commands\n")
	if err := c.SetDefaultBuildImagePrefix(options.NamespaceName, ""); err != nil {
		return err
	}
	return nil
}

func (c *client) applyManifest(manifest *Manifest, options *NamespaceInitOptions, initLabels map[string]string) error {
	for _, release := range manifest.Namespace {
		res, err := manifest.ResourceAbsolutePath(release)
		if err != nil {
			return err
		}
		// Replace any file URL with the corresponding absolute file path.
		absolute, resource, err := furl.IsAbsFile(res)
		if err != nil {
			return err
		}
		if !absolute {
			panic(fmt.Sprintf("manifest.ResourceAbsolutePath returned a non-absolute path: %s", res))
		}

		fmt.Printf("Applying %s in namespace %q\n", release, options.NamespaceName)
		resourceUrl, _ := url.Parse(resource)
		if resourceUrl.Scheme == "" {
			resourceUrl.Scheme = "file"
		}
		labeledContent, err := c.kustomizer.ApplyLabels(resourceUrl, initLabels)
		if err != nil {
			return err
		}
		log, err := c.kubeCtl.ExecStdin([]string{"apply", "-n", options.NamespaceName, "-f", "-"}, &labeledContent)
		fmt.Printf("%s\n", log)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) checkSecretExists(options *NamespaceInitOptions) error {
	_, err := c.kubeClient.CoreV1().Secrets(options.NamespaceName).Get(options.SecretName, v1.GetOptions{})
	return err
}

func getInitLabels() map[string]string {
	return map[string]string{
		"projectriff.io/installer": env.Cli.Name,
		"projectriff.io/version":   env.Cli.Version,
	}
}

func (c *client) appendNewSecretToServiceAccount(namespace, secretName string, serviceAccount *corev1.ServiceAccount) error {
	secretAlreadyPresent := false
	for _, s := range serviceAccount.Secrets {
		if s.Name == secretName {
			secretAlreadyPresent = true
			break
		}
	}
	if secretAlreadyPresent {
		fmt.Printf("Serviceaccount %q already exists in namespace %q with secret %q. Skipping.\n", serviceAccount.Name, namespace, secretName)
		return nil
	}

	serviceAccount.Secrets = append(serviceAccount.Secrets, corev1.ObjectReference{Name: secretName})
	fmt.Printf("Adding secret %q to serviceaccount %q in namespace %q\n", secretName, serviceAccount.Name, namespace)
	_, err := c.kubeClient.CoreV1().ServiceAccounts(namespace).Update(serviceAccount)
	return err
}

func existsSelectors(labelKeys []string) string {
	builder := strings.Builder{}
	for _, key := range labelKeys {
		builder.WriteString(fmt.Sprintf("%s,", key))
	}
	return strings.TrimSuffix(builder.String(), ",")
}

func (c *client) deleteMatchingServiceAccounts(ns string, initLabelSelector string) error {
	serviceAccounts, err := c.kubeClient.CoreV1().ServiceAccounts(ns).List(v1.ListOptions{
		LabelSelector: initLabelSelector,
	})
	if err != nil {
		return err
	}
	deletionResults := tasks.ApplyInParallel(serviceAccountNamesOf(serviceAccounts.Items), func(name string) error {
		return c.kubeClient.CoreV1().ServiceAccounts(ns).Delete(name, &v1.DeleteOptions{})
	})
	return tasks.MergeResults(deletionResults, func(result tasks.CorrelatedResult) string {
		err := result.Error
		if err == nil {
			return ""
		}
		return fmt.Sprintf("Unable to delete service account %s: %v", result.Input, err)
	})
}

func (c *client) deleteMatchingSecrets(ns string, initLabelSelector string) error {
	secrets, err := c.kubeClient.CoreV1().Secrets(ns).List(v1.ListOptions{
		LabelSelector: initLabelSelector,
	})
	if err != nil {
		return err
	}
	deletionResults := tasks.ApplyInParallel(secretNamesOf(secrets.Items), func(name string) error {
		return c.kubeClient.CoreV1().Secrets(ns).Delete(name, &v1.DeleteOptions{})
	})
	return tasks.MergeResults(deletionResults, func(result tasks.CorrelatedResult) string {
		err := result.Error
		if err == nil {
			return ""
		}
		return fmt.Sprintf("Unable to delete secret %s: %v", result.Input, err)
	})
}

func (c *client) deleteMatchingConfigMaps(ns string, initLabelSelector string) error {
	maps, err := c.kubeClient.CoreV1().ConfigMaps(ns).List(v1.ListOptions{
		LabelSelector: initLabelSelector,
	})
	if err != nil {
		return err
	}
	deletionResults := tasks.ApplyInParallel(configMapsNamesOf(maps.Items), func(name string) error {
		return c.kubeClient.CoreV1().ConfigMaps(ns).Delete(name, &v1.DeleteOptions{})
	})
	return tasks.MergeResults(deletionResults, func(result tasks.CorrelatedResult) string {
		err := result.Error
		if err == nil {
			return ""
		}
		return fmt.Sprintf("Unable to delete configmap %s: %v", result.Input, err)
	})
}

func serviceAccountNamesOf(items []corev1.ServiceAccount) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.Name
	}
	return result
}

func secretNamesOf(items []corev1.Secret) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.Name
	}
	return result
}

func configMapsNamesOf(items []corev1.ConfigMap) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.Name
	}
	return result
}

func sortedKeysOf(labels map[string]string) []string {
	result := make([]string, len(labels))
	i := 0
	for key := range labels {
		result[i] = key
		i++
	}
	sort.Strings(result)
	return result
}
