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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pivotal/go-ape/pkg/furl"
	"github.com/projectriff/riff/pkg/env"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	istioNamespace = "istio-system"

	retryDuration = 5 * time.Second
	retrySteps    = 5
)

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

func (c *client) SystemInstall(manifests map[string]*Manifest, options SystemInstallOptions) (bool, error) {
	manifest, err := ResolveManifest(manifests, options.Manifest)
	if err != nil {
		return false, err
	}

	err = ensureNotTerminating(c, allNameSpaces, "Please try again later.")
	if err != nil {
		return false, err
	}

	istioStatus, err := getNamespaceStatus(c, istioNamespace)
	if istioStatus == "'NotFound'" {
		fmt.Print("Installing Istio components\n")
		for i, release := range manifest.Istio {
			if i > 0 {
				time.Sleep(5 * time.Second) // wait for previous resources to be created
			}
			err = c.applyReleaseWithRetry(release, options)
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

	err = waitForIstioComponents(c)
	if err != nil {
		return false, err
	}

	fmt.Print("Installing Knative components\n")
	for _, release := range manifest.Knative {
		err = c.applyReleaseWithRetry(release, options)
		if err != nil {
			return false, err
		}
	}
	fmt.Print("Knative components installed\n\n")
	return true, nil
}

func (c *client) applyReleaseWithRetry(release string, options SystemInstallOptions) error {
	err := retry(retryDuration, retrySteps, func() (bool, error) {
		return c.applyRelease(release, options, true)
	})

	if err != nil {
		// Try again and return the true failure or success.
		_, err = c.applyRelease(release, options, false)
		return err
	}

	return nil
}

var errTimeout = errors.New("timed out")

func retry(duration time.Duration, steps int, condition func() (done bool, err error)) error {
	for i := 0; i < steps; i++ {
		if i != 0 {
			time.Sleep(duration)
		}
		if ok, err := condition(); err != nil || ok {
			return err
		}
	}
	return errTimeout
}

func (c *client) applyRelease(release string, options SystemInstallOptions, willRetry bool) (bool, error) {
	dir, err := furl.Dir(options.Manifest)
	if err != nil {
		return false, err // hard error
	}
	yaml, err := furl.Read(release, dir)
	if err != nil {
		return false, err // hard error
	}
	if options.NodePort {
		yaml = bytes.Replace(yaml, []byte("type: LoadBalancer"), []byte("type: NodePort"), -1)
	}
	fmt.Printf("Applying resources defined in: %s\n", release)
	istioLog, err := c.kubeCtl.ExecStdin([]string{"apply", "-f", "-"}, &yaml)
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
			return false, err // hard error
		}
		if willRetry {
			return false, nil // retriable error
		}
		return false, err // not retrying, so treat as hard error
	}
	return true, nil // success
}

func (c *client) SystemUninstall(options SystemUninstallOptions) (bool, error) {

	err := ensureNotTerminating(c, allNameSpaces, "This would indicate that the system was already uninstalled.")
	if err != nil {
		return false, err
	}
	knativeNsCount, err := checkNamespacesExists(c, knativeNamespaces)
	istioNsCount, err := checkNamespacesExists(c, []string{istioNamespace})
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
		err = deleteKnativeServices(c)
		if err != nil {
			return false, err
		}
		fmt.Print("Removing Knative for " + env.Cli.Name + " components\n")
		err = deleteCrds(c, "knative.dev")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "clusterrolebinding", "knative-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "clusterrolebinding", "build-controller-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "clusterrolebinding", "eventing-controller-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "clusterrolebinding", "in-memory-channel-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "clusterrole", "in-memory-channel-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "clusterrole", "knative-")
		if err != nil {
			return false, err
		}
		deleteSingleResource(c, "service", "knative-ingressgateway", "istio-system")
		deleteSingleResource(c, "horizontalpodautoscaler", "knative-ingressgateway", "istio-system")
		deleteSingleResource(c, "deployment", "knative-ingressgateway", "istio-system")
		err = deleteNamespaces(c, knativeNamespaces)
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
		err = deleteCrds(c, "istio.io")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "clusterrolebinding", "istio-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "clusterrole", "istio-")
		if err != nil {
			return false, err
		}
		err = deleteClusterResources(c, "mutatingwebhookconfiguration", "istio-")
		if err != nil {
			return false, err
		}
		err = deleteNamespaces(c, []string{istioNamespace})
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func waitForIstioComponents(c *client) error {
	fmt.Print("Waiting for the Istio components to start ")
	for i := 0; i < 36; i++ {
		time.Sleep(10 * time.Second) // wait for them to start
		fmt.Print(".")
		pods := c.kubeClient.CoreV1().Pods(istioNamespace)
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

func deleteNamespaces(c *client, namespaces []string) error {
	for _, namespace := range namespaces {
		fmt.Printf("Deleting resources defined in: %s\n", namespace)
		deleteLog, err := c.kubeCtl.Exec([]string{"delete", "namespace", namespace})
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

func deleteSingleResource(c *client, resourceType string, name string, namespace string) error {
	var err error
	var deleteLog string
	if namespace == "" {
		fmt.Printf("Deleting %s/%s resource\n", resourceType, name)
		deleteLog, err = c.kubeCtl.Exec([]string{"delete", resourceType, name})
	} else {
		fmt.Printf("Deleting %s/%s resource in %s\n", resourceType, name, namespace)
		deleteLog, err = c.kubeCtl.Exec([]string{"delete", "-n", namespace, resourceType, name})
	}
	if err != nil {
		if !strings.Contains(deleteLog, "NotFound") {
			fmt.Printf("%s", deleteLog)
		}
	}
	return err
}

func deleteClusterResources(c *client, resourceType string, prefix string) error {
	fmt.Printf("Deleting %ss prefixed with %s\n", resourceType, prefix)
	resourceList, err := c.kubeCtl.Exec([]string{"get", resourceType, "-ocustom-columns=name:metadata.name"})
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
		resourceLog, err := c.kubeCtl.Exec(append([]string{"delete", resourceType}, resourcesToDelete...))
		if err != nil {
			fmt.Printf("%s", resourceLog)
			return err
		}
	}
	return nil
}

func deleteKnativeServices(c *client) error {
	fmt.Printf("Deleting Knative services\n")
	err := deleteAllKnative(c, "service.serving.knative.dev")
	if err != nil {
		return err
	}
	err = deleteAllKnative(c, "configuration.serving.knative.dev")
	if err != nil {
		return err
	}
	err = deleteAllKnative(c, "route.serving.knative.dev")
	if err != nil {
		return err
	}
	return nil
}

func deleteAllKnative(c *client, resourceType string) error {
	resourceList, err := c.kubeCtl.Exec([]string{"get", resourceType, "--all-namespaces", "-ocustom-columns=ns:metadata.namespace,name:metadata.name", "--no-headers=true"})
	if err != nil {
		fmt.Printf("Error while getting %s in all namespaces: %v\n", resourceType, err)
		fmt.Printf("The system seems to be in an unstable state!\n")
	} else {
		resources := strings.Split(string(resourceList), "\n")
		for _, resource := range resources {
			if len(resource) > 0 {
				args := strings.Fields(resource)
				delLog, err := c.kubeCtl.Exec(append([]string{"delete", "-n", args[0], resourceType}, args[1]))
				fmt.Printf("In namespace \"%s\" %s", args[0], delLog)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func deleteCrds(c *client, suffix string) error {
	fmt.Printf("Deleting CRDs for %s\n", suffix)
	crdList, err := c.kubeCtl.Exec([]string{"get", "customresourcedefinitions", "-ocustom-columns=name:metadata.name", "--no-headers=true"})
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
		crdLog, err := c.kubeCtl.Exec(append([]string{"delete", "customresourcedefinition"}, crdsToDelete...))
		if err != nil {
			fmt.Printf("%s", crdLog)
			return err
		}
	}
	return nil
}

func checkNamespacesExists(c *client, names []string) (int, error) {
	count := 0
	for _, name := range names {
		status, err := getNamespaceStatus(c, name)
		if err != nil {
			return count, err
		}
		if status != "'NotFound'" {
			count = +1
		}
	}
	return count, nil
}

func ensureNotTerminating(c *client, names []string, message string) error {
	for _, name := range names {
		status, err := getNamespaceStatus(c, name)
		if err != nil {
			return err
		}
		if status == "'Terminating'" {
			return errors.New(fmt.Sprintf("The %s namespace is currently 'Terminating'. %s", name, message))
		}
	}
	return nil
}

func getNamespaceStatus(c *client, name string) (string, error) {
	nsLog, err := c.kubeCtl.Exec([]string{"get", "namespace", name, "-o", "jsonpath='{.status.phase}'"})
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
