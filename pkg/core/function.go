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
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/boz/kail"
	"k8s.io/apimachinery/pkg/labels"
	"fmt"
	"context"
	"time"
	"github.com/boz/go-logutil"
	"log"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"io"
	"github.com/boz/kcache/types/pod"
	"strings"
	"errors"
	"strconv"
)

const functionLabel = "riff.projectriff.io/function"
const buildAnnotation = "riff.projectriff.io/nonce"

type CreateFunctionOptions struct {
	CreateServiceOptions

	GitRepo     string
	GitRevision string

	InvokerURL string
	Handler    string
	Artifact   string
}

func (c *client) CreateFunction(options CreateFunctionOptions, log io.Writer) (*v1alpha1.Service, error) {
	ns := c.explicitOrConfigNamespace(options.Namespaced)

	s, err := newService(options.CreateServiceOptions)
	if err != nil {
		return nil, err
	}

	labels := s.Spec.RunLatest.Configuration.RevisionTemplate.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labels[functionLabel] = options.Name
	s.Spec.RunLatest.Configuration.RevisionTemplate.SetLabels(labels)
	annotations := s.Spec.RunLatest.Configuration.RevisionTemplate.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[buildAnnotation] = "1"
	s.Spec.RunLatest.Configuration.RevisionTemplate.SetAnnotations(annotations)

	s.Spec.RunLatest.Configuration.Build = &build.BuildSpec{
		ServiceAccountName: "riff-build",
		Source: &build.SourceSpec{
			Git: &build.GitSourceSpec{
				Url:      options.GitRepo,
				Revision: options.GitRevision,
			},
		},
		Template: &build.TemplateInstantiationSpec{
			Name: "riff",
			Arguments: []build.ArgumentSpec{
				{Name: "IMAGE", Value: options.Image},
				{Name: "INVOKER_PATH", Value: options.InvokerURL},
				{Name: "FUNCTION_ARTIFACT", Value: options.Artifact},
				{Name: "FUNCTION_HANDLER", Value: options.Handler},
				{Name: "FUNCTION_NAME", Value: options.Name},
			},
		},
	}

	if !options.DryRun {
		_, err := c.serving.ServingV1alpha1().Services(ns).Create(s)
		if err != nil {
			return nil, err
		}

		if options.Verbose {
			stopChan := make(chan struct{})
			errChan := make(chan error)
			go c.displayFunctionCreationProgress(ns, s.Name, log, stopChan, errChan)
			err := c.waitForSuccessOrFailure(ns, s.Name, stopChan, errChan)
			if err != nil {
				return nil, err
			}
		}
	}

	return s, nil
}

func (c *client) displayFunctionCreationProgress(serviceNamespace string, serviceName string, logWriter io.Writer, stopChan <-chan struct{}, errChan chan<- error) {
	time.Sleep(1000 * time.Millisecond) // ToDo: need some time for revision to get created - is there a better way to slow this down?
	revName, err := c.revisionName(serviceNamespace, serviceName, logWriter, stopChan)
	if err != nil {
		errChan <- err
		return
	} else if revName == "" { // stopped
		return
	}

	ctx := newContext()

	podController, err := c.podController(revName, serviceName, ctx)
	if err != nil {
		errChan <- err
		return
	}

	config, err := c.clientConfig.ClientConfig()
	if err != nil {
		errChan <- err
		return
	}

	controller, err := kail.NewController(ctx, c.kubeClient, config, podController, kail.NewContainerFilter([]string{}), time.Hour)
	if err != nil {
		errChan <- err
		return
	}

	streamLogs(logWriter, controller, stopChan)
	close(errChan)
}

func (c *client) revisionName(serviceNamespace string, serviceName string, logWriter io.Writer, stopChan <-chan struct{}) (string, error) {
	fmt.Fprintf(logWriter, "Waiting for LatestCreatedRevisionName:")
	revName := ""
	for {
		serviceObj, err := c.serving.ServingV1alpha1().Services(serviceNamespace).Get(serviceName, v1.GetOptions{})
		if err != nil {
			return "", err
		}
		revName = serviceObj.Status.LatestCreatedRevisionName
		if revName != "" {
			break
		}
		time.Sleep(1000 * time.Millisecond)
		fmt.Fprintf(logWriter, ".")
		select {
		case <-stopChan:
			return "", nil
		default:
			// continue
		}
	}
	fmt.Fprintf(logWriter, " %s\n", revName)
	return revName, nil
}

func newContext() context.Context {
	ctx := context.Background()
	// avoid kail logs appearing
	l := logutil.New(log.New(ioutil.Discard, "", log.LstdFlags), ioutil.Discard)
	ctx = logutil.NewContext(ctx, l)
	return ctx
}

func (c *client) podController(revName string, serviceName string, ctx context.Context) (pod.Controller, error) {
	dsb := kail.NewDSBuilder()

	buildSelOld, err := labels.Parse(fmt.Sprintf("%s=%s", "build-name", revName))
	if err != nil {
		return nil, err
	}
	buildSel, err := labels.Parse(fmt.Sprintf("%s=%s", "build.knative.dev/buildName", revName))
	if err != nil {
		return nil, err
	}
	runtimeSel, err := labels.Parse(fmt.Sprintf("%s=%s", "serving.knative.dev/configuration", serviceName))
	if err != nil {
		return nil, err
	}
	ds, err := dsb.WithSelectors(or(buildSel, runtimeSel, buildSelOld)).Create(ctx, c.kubeClient) // delete buildSelOld when https://github.com/knative/build/pull/299 is integrated into k/s
	if err != nil {
		return nil, err
	}

	return ds.Pods(), nil
}

func streamLogs(log io.Writer, controller kail.Controller, stopChan <-chan struct{}) {
	events := controller.Events()
	done := controller.Done()
	writer := NewWriter(log)
	for {
		select {
		case ev := <-events:
			// filter out sidecar logs
			container := ev.Source().Container()
			switch container {
			case "queue-proxy":
			case "istio-init":
			case "istio-proxy":
			default:
				writer.Print(ev)
			}
		case <-done:
			return
		case <-stopChan:
			return
		}
	}
}

func (c *client) waitForSuccessOrFailure(namespace string, name string, stopChan chan<- struct{}, errChan <-chan error) error {
	defer close(stopChan)
	for i := 0; i >= 0; i++ {
		select {
		case err := <-errChan:
			return err
		default:
		}
		time.Sleep(500 * time.Millisecond)
		serviceStatusOptions := ServiceStatusOptions{
			Namespaced: Namespaced{namespace},
			Name:       name,
		}
		cond, err := c.ServiceStatus(serviceStatusOptions)
		if err != nil {
			// allow some time for service status to show up
			if i < 20 {
				continue
			}
			return fmt.Errorf("waitForSuccessOrFailure failed to obtain service status: %v", err)
		}

		switch cond.Status {
		case corev1.ConditionTrue:
			return nil
		case corev1.ConditionFalse:
			var message string
			conds, err := c.ServiceConditions(serviceStatusOptions)
			if err != nil {
				// fall back to a basic message
				message = cond.Message
			} else {
				message = serviceConditionsMessage(conds, cond.Message)
			}
			return fmt.Errorf("function create failed: %s: %s", cond.Reason, message)
		default:
			// keep going until outcome is known
		}
	}
	return nil
}

func serviceConditionsMessage(conds []v1alpha1.ServiceCondition,primaryMessage string) string {
	msg := []string{primaryMessage}
	for _, cond := range conds {
		if cond.Status == corev1.ConditionFalse && cond.Type != v1alpha1.ServiceConditionReady && cond.Message != primaryMessage {
			msg = append(msg, cond.Message)
		}
	}
	return strings.Join(msg, "; ")
}

func or(disjuncts ...labels.Selector) labels.Selector {
	return selectorDisjunction(disjuncts)
}

type selectorDisjunction []labels.Selector

func (selectorDisjunction) Add(r ...labels.Requirement) labels.Selector {
	panic("implement me")
}

func (selectorDisjunction) DeepCopySelector() labels.Selector {
	panic("implement me")
}

func (selectorDisjunction) Empty() bool {
	panic("implement me")
}

func (sd selectorDisjunction) Matches(lbls labels.Labels) bool {
	for _, s := range sd {
		if s.Matches(lbls) {
			return true
		}
	}
	return false
}

func (selectorDisjunction) Requirements() (requirements labels.Requirements, selectable bool) {
	panic("implement me")
}

func (selectorDisjunction) String() string {
	panic("implement me")
}

type BuildFunctionOptions struct {
	Namespaced
	Name string
	Verbose bool
}

func (c *client) BuildFunction(options BuildFunctionOptions, log io.Writer) error {
	ns := c.explicitOrConfigNamespace(options.Namespaced)

	s, err := c.service(options.Namespaced, options.Name)
	if err != nil {
		return err
	}

	labels := s.Spec.RunLatest.Configuration.RevisionTemplate.Labels
	if labels[functionLabel] == "" {
		return errors.New(fmt.Sprintf("the service named \"%s\" is not a riff function", options.Name))
	}

	annotations := s.Spec.RunLatest.Configuration.RevisionTemplate.Annotations
	if annotations == nil {
		annotations = map[string]string{}
	}
	build := annotations[buildAnnotation]
	i, err := strconv.Atoi(build)
	if err != nil {
		i = 0
	}
	annotations[buildAnnotation] = strconv.Itoa(i + 1)
	s.Spec.RunLatest.Configuration.RevisionTemplate.SetAnnotations(annotations)

	_, err = c.serving.ServingV1alpha1().Services(s.Namespace).Update(s)
	if err != nil {
		return err
	}

	if options.Verbose {
		stopChan := make(chan struct{})
		errChan := make(chan error)
		go c.displayFunctionCreationProgress(ns, s.Name, log, stopChan, errChan)
		err := c.waitForSuccessOrFailure(ns, s.Name, stopChan, errChan)
		if err != nil {
			return err
		}
	}

	return nil
}


