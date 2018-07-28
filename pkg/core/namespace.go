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

import "fmt"

type Namespaced struct {
	Namespace string
}

type NamespaceInitOptions struct {
	NamespaceName string
	SecretName    string
}

func (c *client) explicitOrConfigNamespace(namespaced Namespaced) string {
	if namespaced.Namespace != "" {
		return namespaced.Namespace
	} else {
		namespace, _, _ := c.clientConfig.Namespace()
		return namespace
	}
}

func (kc *kubectlClient) NamespaceInit(options NamespaceInitOptions) error {

	riffBuildRelease := "https://storage.googleapis.com/riff-releases/riff-build-0.1.1.yaml"

	ns := options.NamespaceName

	fmt.Printf("Initializing %s namespace\n\n", ns)

	if ns != "default" {
		nsYaml := []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: ` + ns)

		nsLog, err := kc.kubeCtl.ExecStdin([]string{"apply", "-f", "-"}, &nsYaml)
		if err != nil {
			print(nsLog)
			print(err.Error(), "\n\n")
		}
	}

	saYaml := []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: riff-build
secrets:
- name: ` + options.SecretName)

	fmt.Printf("Applying serviceaccount resource riff-build using secret %s in namespace %s\n", options.SecretName, ns)
	saLog, err := kc.kubeCtl.ExecStdin([]string{"apply", "-n", ns, "-f", "-"}, &saYaml)
	if err != nil {
		fmt.Print(saLog)
		fmt.Printf("%s\n\n", err.Error())
	} else {
		fmt.Printf("%s\n", saLog)
	}

	riffBuildUrl, err := resolveReleaseURLs(riffBuildRelease)
	if err != nil {
		return err
	}
	fmt.Printf("Applying riff build resources in namespace %s\n", ns)
	riffBuildLog, err := kc.kubeCtl.Exec([]string{"apply", "-f", riffBuildUrl.String()})
	fmt.Printf("%s\n", riffBuildLog)
	if err != nil {
		return err
	}

	return nil
}
