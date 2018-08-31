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
 *
 */

package core

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Namespaced struct {
	Namespace string
}

type NamespaceInitOptions struct {
	NamespaceName string
	SecretName    string
	Manifest      string

	GcrTokenPath      string
	DockerHubUsername string
}

func (c *client) explicitOrConfigNamespace(namespaced Namespaced) string {
	if namespaced.Namespace != "" {
		return namespaced.Namespace
	} else {
		namespace, _, _ := c.clientConfig.Namespace()
		return namespace
	}
}

func (c *kubectlClient) NamespaceInit(options NamespaceInitOptions) error {
	manifest, err := NewManifest(options.Manifest)
	if err != nil {
		return err
	}

	ns := options.NamespaceName

	fmt.Printf("Initializing %s namespace\n\n", ns)

	namespace, err := c.kubeClient.CoreV1().Namespaces().Get(ns, v1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Creating namespace %q \n", ns)
		namespace.Name = ns
		namespace, err = c.kubeClient.CoreV1().Namespaces().Create(namespace)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	secretName := options.SecretName
	if secretName == "" {
		secretName = fmt.Sprintf("push-credentials-%s", randomName())
	}
	if options.GcrTokenPath != "" {
		token, err := ioutil.ReadFile(options.GcrTokenPath)
		if err != nil {
			return err
		}
		secret := &v12.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:        secretName,
				Annotations: map[string]string{"build.knative.dev/docker-0": "https://gcr.io"},
			},
			Type: v12.SecretTypeBasicAuth,
			StringData: map[string]string{
				"username": "_json_key",
				"password": string(token),
			},
		}
		fmt.Printf("Creating secret %q with GCR authentication key from file %s\n", secretName, options.GcrTokenPath)
		_, err = c.kubeClient.CoreV1().Secrets(ns).Create(secret)
		if err != nil {
			return err
		}
	} else if options.DockerHubUsername != "" {
		password, err := readPassword(fmt.Sprintf("Enter dockerhub password for user %q", options.DockerHubUsername))
		if err != nil {
			return err
		}
		secret := &v12.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:        secretName,
				Annotations: map[string]string{"build.knative.dev/docker-0": "https://index.docker.io/v1/"},
			},
			Type: v12.SecretTypeBasicAuth,
			StringData: map[string]string{
				"username": options.DockerHubUsername,
				"password": password,
			},
		}
		fmt.Printf("Creating secret %q with DockerHub authentication for user %q\n", secretName, options.DockerHubUsername)
		_, err = c.kubeClient.CoreV1().Secrets(ns).Create(secret)
		if err != nil {
			return err
		}

	}

	sa := &v12.ServiceAccount{}
	sa.Name = "riff-build"
	sa.Secrets = append(sa.Secrets, v12.ObjectReference{Name: secretName})
	fmt.Printf("Creating serviceaccount %q using secret %q in namespace %q\n", sa.Name, secretName, ns)
	_, err = c.kubeClient.CoreV1().ServiceAccounts(ns).Create(sa)
	if err != nil {
		return err
	}

	for _, release := range manifest.Namespace {
		url, err := resolveReleaseURLs(release)
		if err != nil {
			return err
		}
		fmt.Printf("Applying %s in namespace %q\n", release, ns)
		log, err := c.kubeCtl.Exec([]string{"apply", "-n", ns, "-f", url.String()})
		fmt.Printf("%s\n", log)
		if err != nil {
			return err
		}
	}
	return nil
}

func randomName() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 7)
	for i, _ := range result {
		result[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(result)
}

func readPassword(s string) (string, error) {
	fmt.Print(s)
	if terminal.IsTerminal(int(syscall.Stdin)) {
		res, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")
		return string(res), err
	} else {
		reader := bufio.NewReader(os.Stdin)
		res, err := reader.ReadString('\n')
		return res, err
	}
}
