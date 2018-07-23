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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type SystemInstallOptions struct {
	NodePort bool
}

func (kc *kubectlClient) SystemInstall(options SystemInstallOptions) error {

	istioRelease := "https://storage.googleapis.com/riff-releases/istio-riff-0.1.0.yaml"
	servingRelease := "https://storage.googleapis.com/riff-releases/release-no-mon-riff-0.1.0.yaml"
	eventingRelease := "https://storage.googleapis.com/riff-releases/release-eventing-riff-0.1.0.yaml"
	stubBusRelease := "https://storage.googleapis.com/riff-releases/release-eventing-clusterbus-stub-riff-0.1.0.yaml"

	istioUrl, err := resolveReleaseURLs(istioRelease)
	if err != nil {
		return err
	}
	print("Installing Istio: ", istioUrl.String(), "\n")
	istioYaml, err := loadRelease(istioUrl)
	if err != nil {
		return err
	}
	if options.NodePort {
		istioYaml = bytes.Replace(istioYaml, []byte("LoadBalancer"), []byte("NodePort"), -1)
	}
	istioLog, err := kc.kubeCtl.ExecStdin([]string{"apply", "-f", "-"}, &istioYaml)
	if err != nil {
		print(istioLog, "\n")
		return err
	}
	print("Istio for riff installed\n", "\n")

	err = waitForIstioSidecarInjector(kc)
	if err != nil {
		return err
	}

	servingUrl, err := resolveReleaseURLs(servingRelease)
	if err != nil {
		return err
	}
	print("Installing Knative Serving: ", servingUrl.String(), "\n")
	servingYaml, err := loadRelease(servingUrl)
	if err != nil {
		return err
	}
	if options.NodePort {
		servingYaml = bytes.Replace(servingYaml, []byte("LoadBalancer"), []byte("NodePort"), -1)
	}
	servingLog, err := kc.kubeCtl.ExecStdin([]string{"apply", "-f", "-"}, &servingYaml)
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

func loadRelease(url url.URL) ([]byte, error) {
	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func waitForIstioSidecarInjector(kc *kubectlClient) error {
	print("Waiting for istio-sidecar-injector to start ")
	for i := 0; i < 36; i++ {
		print(".")
		injectorStatus, err := kc.kubeCtl.Exec([]string{"get", "pod", "-n", "istio-system", "-l", "istio=sidecar-injector", "-o", "jsonpath='{.items[0].status.phase}'"})
		if err != nil {
			return err
		}
		if injectorStatus == "'Error'" {
			return errors.New("istio-sidecar-injector pod failed to start")
		}
		if injectorStatus == "'Running'" {
			print(injectorStatus, "\n\n")
			return nil
		}
		time.Sleep(10 * time.Second) // wait for it to start
	}
	print("\n\n")
	return errors.New("istio-sidecar-injector pod did not start in time")
}
