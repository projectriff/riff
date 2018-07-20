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

	buildtemplateRelease := "https://storage.googleapis.com/riff-releases/riff-buildtemplate-0.1.0.yaml"

	ns := options.NamespaceName

	print("Initializing ", ns, " namespace ", "\n\n")

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

	print("Labeling namespace ", ns, " for sidecar injection", "\n")
	cmdArgs := []string{"label", "namespace", ns, "istio-injection=enabled"}
	injectLog, err := kc.kubeCtl.Exec(cmdArgs)
	if err != nil {
		print("namespace ", ns, " already labeled", "\n\n")
	} else {
		print(injectLog, "\n")
	}

	saYaml := []byte(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: riff-build
secrets:
- name: ` + options.SecretName)

	print("Applying serviceaccount resource riff-build using secret ", options.SecretName, " in namespace ", ns, "\n")
	saLog, err := kc.kubeCtl.ExecStdin([]string{"apply", "-n", ns, "-f", "-"}, &saYaml)
	if err != nil {
		print(saLog)
		print(err.Error(), "\n\n")
	} else {
		print(saLog, "\n")
	}

	buildtemplateUrl, err := resolveReleaseURLs(buildtemplateRelease)
	if err != nil {
		return err
	}
	print("Applying buildtemplate resource riff in namespace ", ns, "\n")
	buildtemplateLog, err := kc.kubeCtl.Exec([]string{"apply", "-f", buildtemplateUrl.String()})
	print(buildtemplateLog, "\n")
	if err != nil {
		return err
	}

	return nil
}

