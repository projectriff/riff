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
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/projectriff/riff/pkg/fileutils"

	"github.com/projectriff/riff/pkg/env"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const istioNamespace = "istio-system"

type SystemInstallOptions struct {
	Manifest string
	NodePort bool
	Force    bool
}

type SystemUninstallOptions struct {
	Istio bool
	Force bool
}

var (
	knativeNamespaces = []string{"knative-eventing", "knative-serving", "knative-build", "knative-monitoring"}
	allNameSpaces     = append(knativeNamespaces, istioNamespace)
)

func (kc *kubectlClient) SystemInstall(manifests map[string]*Manifest, options SystemInstallOptions) (bool, error) {
	manifest, err := ResolveManifest(manifests, options.Manifest)
	if err != nil {
		return false, err
	}

	err = ensureNotTerminating(kc, allNameSpaces, "Please try again later.")
	if err != nil {
		return false, err
	}

	istioStatus, err := getNamespaceStatus(kc, istioNamespace)
	if istioStatus == "'NotFound'" {
		fmt.Print("Installing Istio components\n")
		for i, release := range manifest.Istio {
			if i > 0 {
				time.Sleep(5 * time.Second) // wait for previous resources to be created
			}
			err = kc.applyReleaseWithRetry(release, options)
			if err != nil {
				return false, err
			}
		}
		fmt.Print("Istio components installed\n\n")
	} else {
		if !options.Force {
			answer, err := confirm("Istio is already installed, do you want to install the Knative components for riff?")
			if err != nil {
				return false, err
			}
			if !answer {
				return false, nil
			}
		}
	}

	err = waitForIstioComponents(kc)
	if err != nil {
		return false, err
	}

	fmt.Print("Installing Knative components\n")
	for _, release := range manifest.Knative {
		err = kc.applyReleaseWithRetry(release, options)
		if err != nil {
			return false, err
		}
	}
	fmt.Print("Knative components installed\n\n")
	return true, nil
}

func (kc *kubectlClient) applyReleaseWithRetry(release string, options SystemInstallOptions) error {
	err := kc.applyRelease(release, options)
	if err != nil {
		fmt.Printf("Error applying resources, trying again\n")
		time.Sleep(5 * time.Second) // wait for previous resources to be created
		return kc.applyRelease(release, options)
	}
	return nil
}

func (kc *kubectlClient) applyRelease(release string, options SystemInstallOptions) error {
	dir, err := fileutils.Dir(options.Manifest)
	if err != nil {
		return err
	}
	yaml, err := fileutils.Read(release, dir)
	if err != nil {
		return err
	}
	if options.NodePort {
		yaml = bytes.Replace(yaml, []byte("type: LoadBalancer"), []byte("type: NodePort"), -1)
	}
	fmt.Printf("Applying resources defined in: %s\n", release)
	istioLog, err := kc.kubeCtl.ExecStdin([]string{"apply", "-f", "-"}, &yaml)
	if err != nil {
		fmt.Printf("%s\n", istioLog)
		if strings.Contains(istioLog, "forbidden") {
			fmt.Print(`It looks like you don't have cluster-admin permissions.

To fix this you need to:
 1. Delete the current failed installation using:
      ` + env.Cli.Name + ` system uninstall --istio --force
 2. Give the user account used for installation cluster-admin permissions, you can use the following command:
      kubectl create clusterrolebinding cluster-admin-binding \
        --clusterrole=cluster-admin \
        --user=<install-user>
 3. Re-install ` + env.Cli.Name + `

`)
		}
		return err
	}
	return nil
}

func (kc *kubectlClient) SystemUninstall(options SystemUninstallOptions) (bool, error) {

	err := ensureNotTerminating(kc, allNameSpaces, "This would indicate that the system was already uninstalled.")
	if err != nil {
		return false, err
	}
	knativeNsCount, err := checkNamespacesExists(kc, knativeNamespaces)
	istioNsCount, err := checkNamespacesExists(kc, []string{istioNamespace})
	if err != nil {
		return false, err
	}
	if knativeNsCount == 0 {
		fmt.Print("No Knative components for " + env.Cli.Name + " found\n")
	} else {
		if !options.Force {
			answer, err := confirm("Are you sure you want to uninstall the " + env.Cli.Name + " system?")
			if err != nil {
				return false, err
			}
			if !answer {
				return false, nil
			}
		}
		fmt.Print("Removing Knative for " + env.Cli.Name + " components\n")
		err = deleteCrds(kc, "knative.dev")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(kc, "clusterrolebinding", "knative-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(kc, "clusterrolebinding", "build-controller-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(kc, "clusterrolebinding", "eventing-controller-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(kc, "clusterrolebinding", "in-memory-channel-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(kc, "clusterrole", "in-memory-channel-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(kc, "clusterrole", "knative-")
		if err != nil {
			return false, err
		}
		deleteSingleResource(kc, "service", "knative-ingressgateway", "istio-system")
		deleteSingleResource(kc, "horizontalpodautoscaler", "knative-ingressgateway", "istio-system")
		deleteSingleResource(kc, "deployment", "knative-ingressgateway", "istio-system")
		err = deleteNamespaces(kc, knativeNamespaces)
		if err != nil {
			return false, err
		}
	}
	if istioNsCount == 0 {
		fmt.Print("No Istio components found\n")
	} else {
		if !options.Istio {
			if options.Force {
				return true, nil
			}
			answer, err := confirm("Do you also want to uninstall Istio components?")
			if err != nil {
				return false, err
			}
			if !answer {
				return false, nil
			}
		}
		fmt.Print("Removing Istio components\n")
		err = deleteCrds(kc, "istio.io")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(kc, "clusterrolebinding", "istio-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(kc, "clusterrole", "istio-")
		if err != nil {
			return false, err
		}
		err = deleteNamespaces(kc, []string{istioNamespace})
		if err != nil {
			return false, err
		}
		// TODO: remove this once https://github.com/knative/serving/issues/2018 is resolved
		deleteSingleResource(kc, "horizontalpodautoscaler.autoscaling", "istio-pilot", "")
	}
	return true, nil
}

func waitForIstioComponents(kc *kubectlClient) error {
	fmt.Print("Waiting for the Istio components to start ")
	for i := 0; i < 36; i++ {
		time.Sleep(10 * time.Second) // wait for them to start
		fmt.Print(".")
		pods := kc.kubeClient.CoreV1().Pods(istioNamespace)
		podList, err := pods.List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		if len(podList.Items) < 3 {
			// make sure we found pods and not that the system is slow to create pods
			continue
		}
		waitLonger := false
		for _, pod := range podList.Items {
			if !strings.HasPrefix(pod.Name, "istio-") {
				continue
			}
			if pod.Status.Phase != "Running" && pod.Status.Phase != "Succeeded" {
				waitLonger = true
				break
			} else {
				if pod.Status.Phase == "Running" {
					containers := pod.Status.ContainerStatuses
					for _, cont := range containers {
						if !cont.Ready {
							waitLonger = true
							break
						}
					}
				}
			}
		}
		if !waitLonger {
			fmt.Print(" all components are 'Running'\n\n")
			return nil
		}
	}
	return errors.New("the Istio components did not start in time")
}

func deleteNamespaces(kc *kubectlClient, namespaces []string) error {
	for _, namespace := range namespaces {
		fmt.Printf("Deleting resources defined in: %s\n", namespace)
		deleteLog, err := kc.kubeCtl.Exec([]string{"delete", "namespace", namespace})
		if err != nil {
			if strings.Contains(deleteLog, "NotFound") {
				fmt.Printf("Namespace \"%s\" was not found\n", namespace)
			} else {
				fmt.Printf("%s", deleteLog)
			}
		}
	}
	return nil
}

func deleteSingleResource(kc *kubectlClient, resourceType string, name string, namespace string) error {
	var err error
	var deleteLog string
	if namespace == "" {
		fmt.Printf("Deleting %s/%s resource\n", resourceType, name)
		deleteLog, err = kc.kubeCtl.Exec([]string{"delete", resourceType, name})
	} else {
		fmt.Printf("Deleting %s/%s resource in %s\n", resourceType, name, namespace)
		deleteLog, err = kc.kubeCtl.Exec([]string{"delete", "-n", namespace, resourceType, name})
	}
	if err != nil {
		if !strings.Contains(deleteLog, "NotFound") {
			fmt.Printf("%s", deleteLog)
		}
	}
	return err
}

func deleteClusterResources(kc *kubectlClient, resourceType string, prefix string) error {
	fmt.Printf("Deleting %ss prefixed with %s\n", resourceType, prefix)
	resourceList, err := kc.kubeCtl.Exec([]string{"get", resourceType, "-ocustom-columns=name:metadata.name"})
	if err != nil {
		return err
	}
	resource := strings.Split(string(resourceList), "\n")
	var resourcesToDelete []string
	for _, resource := range resource {
		if strings.HasPrefix(resource, prefix) {
			resourcesToDelete = append(resourcesToDelete, resource)
		}
	}
	if len(resourcesToDelete) > 0 {
		resourceLog, err := kc.kubeCtl.Exec(append([]string{"delete", resourceType}, resourcesToDelete...))
		if err != nil {
			fmt.Printf("%s", resourceLog)
			return err
		}
	}
	return nil
}

func deleteCrds(kc *kubectlClient, suffix string) error {
	fmt.Printf("Deleting CRDs for %s\n", suffix)
	crdList, err := kc.kubeCtl.Exec([]string{"get", "customresourcedefinitions", "-ocustom-columns=name:metadata.name"})
	if err != nil {
		return err
	}
	crds := strings.Split(string(crdList), "\n")
	var crdsToDelete []string
	for _, crd := range crds {
		if strings.HasSuffix(crd, suffix) {
			crdsToDelete = append(crdsToDelete, crd)
		}
	}
	if len(crdsToDelete) > 0 {
		crdLog, err := kc.kubeCtl.Exec(append([]string{"delete", "customresourcedefinition"}, crdsToDelete...))
		if err != nil {
			fmt.Printf("%s", crdLog)
			return err
		}
	}
	return nil
}

func checkNamespacesExists(kc *kubectlClient, names []string) (int, error) {
	count := 0
	for _, name := range names {
		status, err := getNamespaceStatus(kc, name)
		if err != nil {
			return count, err
		}
		if status != "'NotFound'" {
			count = +1
		}
	}
	return count, nil
}

func ensureNotTerminating(kc *kubectlClient, names []string, message string) error {
	for _, name := range names {
		status, err := getNamespaceStatus(kc, name)
		if err != nil {
			return err
		}
		if status == "'Terminating'" {
			return errors.New(fmt.Sprintf("The %s namespace is currently 'Terminating'. %s", name, message))
		}
	}
	return nil
}

func getNamespaceStatus(kc *kubectlClient, name string) (string, error) {
	nsLog, err := kc.kubeCtl.Exec([]string{"get", "namespace", name, "-o", "jsonpath='{.status.phase}'"})
	if err != nil {
		if strings.Contains(nsLog, "NotFound") {
			return "'NotFound'", nil
		}
		return "", err
	}
	return nsLog, nil
}

func confirm(s string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", s)
	res, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	if len(res) < 2 {
		return false, nil
	}
	answer := strings.ToLower(strings.TrimSpace(res))[0] == 'y'
	return answer, nil
}
