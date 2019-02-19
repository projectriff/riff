/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"bufio"
	"fmt"
	"github.com/projectriff/riff/pkg/core/tasks"
	"github.com/projectriff/riff/pkg/env"
	"github.com/projectriff/riff/pkg/fileutils"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net/url"
	"os"
	"sort"
	"strings"
	"syscall"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const serviceAccountName = "riff-build"

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

	NoSecret          bool
	SecretName        string
	GcrTokenPath      string
	DockerHubUsername string

	RegistryProtocol string
	RegistryHost     string
	RegistryUser     string
}

type NamespaceCleanupOptions struct {
	NamespaceName   string
	SecretName      string
	RemoveNamespace bool
}

func (o *NamespaceInitOptions) secretType() secretType {
	switch {
	case o.NoSecret:
		return secretTypeNone
	case o.DockerHubUsername != "":
		return secretTypeDockerHub
	case o.GcrTokenPath != "":
		return secretTypeGcr
	case o.RegistryHost != "":
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

func (kc *kubectlClient) NamespaceInit(manifests map[string]*Manifest, options NamespaceInitOptions) error {
	manifest, err := ResolveManifest(manifests, options.Manifest)
	if err != nil {
		return err
	}

	ns := options.NamespaceName

	fmt.Printf("Initializing namespace %q\n\n", ns)

	namespace, err := kc.kubeClient.CoreV1().Namespaces().Get(ns, v1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Creating namespace %q \n", ns)
		namespace = &corev1.Namespace{}
		namespace.Name = ns
		namespace, err = kc.kubeClient.CoreV1().Namespaces().Create(namespace)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	initLabels := getInitLabels()

	switch options.secretType() {
	case secretTypeGcr:
		if err := kc.createGcrSecret(options, initLabels); err != nil {
			return err
		}
	case secretTypeDockerHub:
		if err := kc.createDockerHubSecret(options, initLabels); err != nil {
			return err
		}
	case secretTypeBasicAuth:
		if err := kc.createRegistrySecret(options, initLabels); err != nil {
			return err
		}
	case secretTypeUserProvided:
		if err = kc.checkSecretExists(options); err != nil {
			return err
		}

	}

	sa, err := kc.kubeClient.CoreV1().ServiceAccounts(ns).Get(serviceAccountName, v1.GetOptions{})
	if errors.IsNotFound(err) {
		sa = &corev1.ServiceAccount{}
		sa.Name = serviceAccountName
		sa.Labels = initLabels
		if options.secretType() != secretTypeNone {
			secretName := options.SecretName
			sa.Secrets = append(sa.Secrets, corev1.ObjectReference{Name: secretName})
			fmt.Printf("Creating serviceaccount %q using secret %q in namespace %q\n", sa.Name, secretName, ns)
		} else {
			fmt.Printf("Creating unauthenticated serviceaccount %q in namespace %q\n", sa.Name, ns)
		}
		_, err = kc.kubeClient.CoreV1().ServiceAccounts(ns).Create(sa)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if options.secretType() != secretTypeNone {
		secretName := options.SecretName
		secretAlreadyPresent := false
		for _, s := range sa.Secrets {
			if s.Name == secretName {
				secretAlreadyPresent = true
				break
			}
		}
		if secretAlreadyPresent {
			fmt.Printf("Serviceaccount %q already exists in namespace %q with secret %q. Skipping.\n", sa.Name, ns, secretName)
		} else {
			sa.Secrets = append(sa.Secrets, corev1.ObjectReference{Name: secretName})
			fmt.Printf("Adding secret %q to serviceaccount %q in namespace %q\n", secretName, sa.Name, ns)
			_, err = kc.kubeClient.CoreV1().ServiceAccounts(ns).Update(sa)
			if err != nil {
				return err
			}
		}
	}

	for _, release := range manifest.Namespace {
		res, err := manifest.ResourceAbsolutePath(release)
		if err != nil {
			return err
		}

		// Replace any file URL with the corresponding absolute file path.
		absolute, resource, err := fileutils.IsAbsFile(res)
		if err != nil {
			return err
		}
		if !absolute {
			panic(fmt.Sprintf("manifest.ResourceAbsolutePath returned a non-absolute path: %s", res))
		}

		fmt.Printf("Applying %s in namespace %q\n", release, ns)
		resourceUrl, _ := url.Parse(resource)
		if resourceUrl.Scheme == "" {
			resourceUrl.Scheme = "file"
		}
		labeledContent, err := kc.kustomizer.ApplyLabels(resourceUrl, initLabels)
		if err != nil {
			return err
		}
		log, err := kc.kubeCtl.ExecStdin([]string{"apply", "-n", ns, "-f", "-"}, &labeledContent)
		fmt.Printf("%s\n", log)
		if err != nil {
			return err
		}
	}
	return nil
}

func (kc *kubectlClient) NamespaceCleanup(options NamespaceCleanupOptions) error {
	ns := options.NamespaceName
	initLabelKeys := sortedKeysOf(getInitLabels())
	initLabelSelector := existsSelectors(initLabelKeys)

	fmt.Printf("Deleting serviceaccounts matching label keys %v in namespace %q\n", initLabelKeys, ns)
	if err := kc.deleteMatchingServiceAccounts(ns, initLabelSelector); err != nil {
		return err
	}

	fmt.Printf("Deleting persistentvolumeclaims matching label keys %v in namespace %q\n", initLabelKeys, ns)
	if err := kc.deleteMatchingPersistentVolumeClaims(ns, initLabelSelector); err != nil {
		return err
	}

	fmt.Printf("Deleting secrets matching label keys %v in namespace %q\n", initLabelKeys, ns)
	if err := kc.deleteMatchingSecrets(ns, initLabelSelector); err != nil {
		return err
	}

	if options.RemoveNamespace {
		fmt.Printf("Deleting namespace %q\n", ns)
		if err := kc.kubeClient.CoreV1().Namespaces().Delete(ns, &v1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (kc *kubectlClient) checkSecretExists(options NamespaceInitOptions) error {
	_, err := kc.kubeClient.CoreV1().Secrets(options.NamespaceName).Get(options.SecretName, v1.GetOptions{})
	return err
}

func (kc *kubectlClient) createDockerHubSecret(options NamespaceInitOptions, labels map[string]string) error {
	username := options.DockerHubUsername
	password, err := readPassword(fmt.Sprintf("Enter password for user %q", username))
	if err != nil {
		return err
	}
	return kc.createBasicAuthSecret(options.NamespaceName, options.SecretName, username, password, "https://index.docker.io/v1/", labels)
}

func (kc *kubectlClient) createGcrSecret(options NamespaceInitOptions, labels map[string]string) error {
	token, err := ioutil.ReadFile(options.GcrTokenPath)
	if err != nil {
		return err
	}
	return kc.createBasicAuthSecret(options.NamespaceName, options.SecretName, "_json_key", string(token), "https://gcr.io", labels)
}

func (kc *kubectlClient) createRegistrySecret(options NamespaceInitOptions, labels map[string]string) error {
	username := options.RegistryUser
	password, err := readPassword(fmt.Sprintf("Enter password for user %q", username))
	if err != nil {
		return err
	}
	registryAddress := fmt.Sprintf("%s://%s", options.RegistryProtocol, options.RegistryHost)
	return kc.createBasicAuthSecret(options.NamespaceName, options.SecretName, username, password, registryAddress, labels)
}

func (kc *kubectlClient) createBasicAuthSecret(namespace string,
	secretName string,
	username string,
	password string,
	serverAddress string,
	initLabels map[string]string) error {

	_ = kc.kubeClient.CoreV1().Secrets(namespace).Delete(secretName, &v1.DeleteOptions{})

	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:        secretName,
			Annotations: map[string]string{"build.knative.dev/docker-0": serverAddress},
			Labels:      initLabels,
		},
		Type: corev1.SecretTypeBasicAuth,
		StringData: map[string]string{
			"username": username,
			"password": password,
		},
	}
	fmt.Printf("Creating secret %q with basic authentication to server %q for user %q\n", secretName, serverAddress, username)

	_, err := kc.kubeClient.CoreV1().Secrets(namespace).Create(secret)
	return err
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	if terminal.IsTerminal(int(syscall.Stdin)) {
		res, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")
		return string(res), err
	} else {
		reader := bufio.NewReader(os.Stdin)
		res, err := ioutil.ReadAll(reader)
		return string(res), err
	}
}

func getInitLabels() map[string]string {
	return map[string]string{
		"projectriff.io/installer": env.Cli.Name,
		"projectriff.io/version":   env.Cli.Version,
	}
}

func existsSelectors(labelKeys []string) string {
	builder := strings.Builder{}
	for _, key := range labelKeys {
		builder.WriteString(fmt.Sprintf("%s,", key))
	}
	return strings.TrimSuffix(builder.String(), ",")
}

func (kc *kubectlClient) deleteMatchingServiceAccounts(ns string, initLabelSelector string) error {
	serviceAccounts, err := kc.kubeClient.CoreV1().ServiceAccounts(ns).List(v1.ListOptions{
		LabelSelector: initLabelSelector,
	})
	if err != nil {
		return err
	}
	deletionResults := tasks.ApplyInParallel(serviceAccountNamesOf(serviceAccounts.Items), func(name string) error {
		return kc.kubeClient.CoreV1().ServiceAccounts(ns).Delete(name, &v1.DeleteOptions{})
	})
	return tasks.MergeResults(deletionResults, func(result tasks.CorrelatedResult) string {
		err := result.Error
		if err == nil {
			return ""
		}
		return fmt.Sprintf("Unable to delete service account %s: %v", result.Input, err)
	})
}

func (kc *kubectlClient) deleteMatchingPersistentVolumeClaims(ns string, initLabelSelector string) error {
	persistentVolumeClaims, err := kc.kubeClient.CoreV1().PersistentVolumeClaims(ns).List(v1.ListOptions{
		LabelSelector: initLabelSelector,
	})
	if err != nil {
		return err
	}
	deletionResults := tasks.ApplyInParallel(persistentVolumeClaimNamesOf(persistentVolumeClaims.Items), func(name string) error {
		return kc.kubeClient.CoreV1().PersistentVolumeClaims(ns).Delete(name, &v1.DeleteOptions{})
	})
	return tasks.MergeResults(deletionResults, func(result tasks.CorrelatedResult) string {
		err := result.Error
		if err == nil {
			return ""
		}
		return fmt.Sprintf("Unable to delete persistent volume claim %s: %v", result.Input, err)
	})
}

func (kc *kubectlClient) deleteMatchingSecrets(ns string, initLabelSelector string) error {
	secrets, err := kc.kubeClient.CoreV1().Secrets(ns).List(v1.ListOptions{
		LabelSelector: initLabelSelector,
	})
	if err != nil {
		return err
	}
	deletionResults := tasks.ApplyInParallel(secretNamesOf(secrets.Items), func(name string) error {
		return kc.kubeClient.CoreV1().Secrets(ns).Delete(name, &v1.DeleteOptions{})
	})
	return tasks.MergeResults(deletionResults, func(result tasks.CorrelatedResult) string {
		err := result.Error
		if err == nil {
			return ""
		}
		return fmt.Sprintf("Unable to delete secret %s: %v", result.Input, err)
	})
}

func serviceAccountNamesOf(items []corev1.ServiceAccount) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.Name
	}
	return result
}

func persistentVolumeClaimNamesOf(items []corev1.PersistentVolumeClaim) []string {
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
