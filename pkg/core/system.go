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
	"net/url"
	"fmt"
)

type SystemInstallOptions struct {
	NodePort bool
}

func (kc *kubectlClient) SystemInstall(options SystemInstallOptions) error {

	istioRelease := "https://storage.googleapis.com/riff-releases/istio-riff-0.1.0.yaml"
	servingRelease := "https://storage.googleapis.com/knative-releases/latest/release-no-mon.yaml"
	eventingRelease := "https://storage.googleapis.com/knative-releases/latest/release-eventing.yaml"
	stubBusRelease := "https://storage.googleapis.com/knative-releases/latest/release-eventing-clusterbus-stub.yaml"

	if options.NodePort {
		print("The --node-port option is not supported yet\n")
	}

	istioUrl, err := resolveReleaseURLs(istioRelease)
	if err != nil {
		return err
	}
	print("Installing Istio: ", istioUrl.String(), "\n")
	istioLog, err := kc.kubeCtl.Exec([]string{"apply", "-f", istioUrl.String()})
	if err != nil {
		print(istioLog, "\n")
		return err
	}
	print("Istio for riff installed\n", "\n")

	servingUrl, err := resolveReleaseURLs(servingRelease)
	if err != nil {
		return err
	}
	print("Installing Knative Serving: ", servingUrl.String(), "\n")
	servingLog, err := kc.kubeCtl.Exec([]string{"apply", "-f", servingUrl.String()})
	if err != nil {
		print(servingLog, "\n")
		return err
	}
	print("Knative Serving for riff installed\n", "\n")

	eventingUrl, err := resolveReleaseURLs(eventingRelease)
	if err != nil {
		return err
	}
	print("Installing Knative Eventing: ", eventingUrl.String(), "\n")
	eventingLog, err := kc.kubeCtl.Exec([]string{"apply", "-f", eventingUrl.String()})
	if err != nil {
		print(eventingLog, "\n")
		return err
	}
	print("Knative Eventing for riff installed\n", "\n")

	busUrl, err := resolveReleaseURLs(stubBusRelease)
	if err != nil {
		return err
	}
	print("Applying Stub ClusterBus resource: ", busUrl.String(), "\n")
	busLog, err := kc.kubeCtl.Exec([]string{"apply", "-f", busUrl.String()})
	print(busLog, "\n")
	if err != nil {
		return err
	}

	print("riff system install is complete\n", "\n")

	return nil
}

func resolveReleaseURLs(filename string) (url.URL, error) {
	u, err := url.Parse(filename)
	if err != nil {
		return url.URL{}, err
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return *u, nil
	}
	return *u, fmt.Errorf("Filename must be file, http or https, got %s", u.Scheme)
}
