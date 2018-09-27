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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/boz/go-logutil"
	"github.com/boz/kail"
	"github.com/boz/kcache/types/pod"
	"github.com/buildpack/pack"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	functionLabel   = "riff.projectriff.io/function"
	buildAnnotation = "riff.projectriff.io/nonce"
	// annotation names with slashes are rejected :-/
	buildpackBuildImageAnnotation = "riff.projectriff.io-buildpack-buildImage"
	buildpackRunImageAnnotation   = "riff.projectriff.io-buildpack-runImage"
)

type CreateFunctionOptions struct {
	CreateOrReviseServiceOptions

	LocalPath   string
	GitRepo     string
	GitRevision string

	Invoker        string
	BuildpackImage string
	InvokerURL     string

	Handler  string
	Artifact string
}

func (c *client) CreateFunction(options CreateFunctionOptions, log io.Writer) (*v1alpha1.Service, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	s, err := newService(options.CreateOrReviseServiceOptions)
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

	if options.InvokerURL != "" {
		if options.LocalPath != "" {
			return nil, fmt.Errorf("the selected invoker %q does not support local builds", options.Invoker)
		}
		// invoker based cluster build
		s.Spec.RunLatest.Configuration.Build = &build.BuildSpec{
			ServiceAccountName: "riff-build",
			Source:             c.makeBuildSourceSpec(options),
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
	} else if options.BuildpackImage != "" {
		// TODO support options.Artifact and options.Handler
		if options.LocalPath != "" {
			appDir := options.LocalPath
			buildImage := options.BuildpackImage
			runImage := "packs/run"
			repoName := options.Image
			publish := publishImage(repoName)

			if s.ObjectMeta.Annotations == nil {
				s.ObjectMeta.Annotations = make(map[string]string)
			}
			s.ObjectMeta.Annotations[buildpackBuildImageAnnotation] = buildImage
			s.ObjectMeta.Annotations[buildpackRunImageAnnotation] = runImage

			if options.DryRun {
				// skip build for a dry run
				log.Write([]byte("Skipping local build\n"))
			} else {
				err := pack.Build(appDir, buildImage, runImage, repoName, publish)
				if err != nil {
					return nil, err
				}
			}
		} else {
			// buildpack based cluster build
			s.Spec.RunLatest.Configuration.Build = &build.BuildSpec{
				ServiceAccountName: "riff-build",
				Source:             c.makeBuildSourceSpec(options),
				Template: &build.TemplateInstantiationSpec{
					Name: "riff-cnb",
					Arguments: []build.ArgumentSpec{
						{Name: "IMAGE", Value: options.Image},
						// TODO configure buildtemplate based on buildpack image
						// {Name: "TBD", Value: options.BuildpackImage},
					},
				},
			}
		}
	} else {
		return nil, fmt.Errorf("unknown build permutation")
	}

	if !options.DryRun {
		_, err := c.serving.ServingV1alpha1().Services(ns).Create(s)
		if err != nil {
			return nil, err
		}

		if options.Verbose || options.Wait {
			stopChan := make(chan struct{})
			errChan := make(chan error)
			if options.Verbose {
				go c.displayFunctionCreationProgress(ns, s.Name, log, stopChan, errChan)
			}
			err := c.waitForSuccessOrFailure(ns, s.Name, 1, stopChan, errChan)
			if err != nil {
				return nil, err
			}
		}
	}

	return s, nil
}

func (c *client) makeBuildSourceSpec(options CreateFunctionOptions) *build.SourceSpec {
	return &build.SourceSpec{
		Git: &build.GitSourceSpec{
			Url:      options.GitRepo,
			Revision: options.GitRevision,
		},
	}
}

func (c *client) displayFunctionCreationProgress(serviceNamespace string, serviceName string, logWriter io.Writer, stopChan <-chan struct{}, errChan chan<- error) {
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

func (c *client) waitForSuccessOrFailure(namespace string, name string, gen int64, stopChan chan<- struct{}, errChan <-chan error) error {
	defer close(stopChan)
	for i := 0; i >= 0; i++ {
		select {
		case err := <-errChan:
			return err
		default:
		}
		time.Sleep(500 * time.Millisecond)
		service, err := c.service(namespace, name)
		if err != nil {
			return fmt.Errorf("waitForSuccessOrFailure failed to obtain service: %v", err)
		}
		if service.Status.ObservedGeneration < gen {
			// allow some time for service status observed generation to show up
			if i < 20 {
				continue
			}
			return fmt.Errorf("waitForSuccessOrFailure failed to obtain service status for observedGeneration %d: %v", gen, err)
		}
		serviceStatusOptions := ServiceStatusOptions{
			Namespace: namespace,
			Name:      name,
		}
		cond, err := c.ServiceStatus(serviceStatusOptions)
		if err != nil {
			return fmt.Errorf("waitForSuccessOrFailure failed to obtain service status: %v", err)
		}

		switch cond.Status {
		case corev1.ConditionTrue:
			return nil
		case corev1.ConditionFalse:
			someStateIsUnknown := false
			conds, err := c.ServiceConditions(serviceStatusOptions)
			if err == nil {
				for _, c := range conds {
					if c.Status == corev1.ConditionUnknown {
						someStateIsUnknown = true
						break
					}
				}
			}
			if !someStateIsUnknown {
				var message string
				if err != nil {
					// fall back to a basic message
					message = cond.Message
				} else {
					message = serviceConditionsMessage(conds, cond.Message)
				}
				return fmt.Errorf("function creation failed: %s: %s", cond.Reason, message)
			}
		default:
			// keep going until outcome is known
		}
	}
	return nil
}

func serviceConditionsMessage(conds []v1alpha1.ServiceCondition, primaryMessage string) string {
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
	Namespace string
	Name      string
	LocalPath string
	Verbose   bool
	Wait      bool
}

func (c *client) getServiceSpecGeneration(namespace string, name string) (int64, error) {
	s, err := c.service(namespace, name)
	if err != nil {
		return 0, err
	}
	return s.Spec.Generation, nil
}

func (c *client) BuildFunction(options BuildFunctionOptions, log io.Writer) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	service, err := c.service(options.Namespace, options.Name)
	if err != nil {
		return err
	}

	// create a copy before mutating
	service = service.DeepCopy()

	gen := service.Spec.Generation

	// TODO support non-RunLatest configurations
	configuration := service.Spec.RunLatest.Configuration
	build := configuration.Build
	annotations := service.Annotations
	labels := configuration.RevisionTemplate.Labels
	if labels[functionLabel] == "" {
		return fmt.Errorf("the service named \"%s\" is not a riff function", options.Name)
	}

	c.bumpNonceAnnotation(service)

	if build == nil {
		// function was build locally, attempt to reconstruct configuration
		appDir := options.LocalPath
		buildImage := annotations[buildpackBuildImageAnnotation]
		runImage := annotations[buildpackRunImageAnnotation]
		repoName := configuration.RevisionTemplate.Spec.Container.Image
		publish := publishImage(repoName)

		if buildImage == "" || runImage == "" {
			return fmt.Errorf("unable to build function locally not built from a buildpack")
		}
		if appDir == "" {
			return fmt.Errorf("local-path must be specified to rebuild function from source")
		}

		err := pack.Build(appDir, buildImage, runImage, repoName, publish)
		if err != nil {
			return err
		}
	}

	_, err = c.serving.ServingV1alpha1().Services(service.Namespace).Update(service)
	if err != nil {
		return err
	}

	if options.Verbose || options.Wait {
		stopChan := make(chan struct{})
		errChan := make(chan error)
		var (
			nextGen int64
			err     error
		)
		for i := 0; i < 10; i++ {
			if i >= 10 {
				return fmt.Errorf("build unsuccesful for \"%s\", service resource was never updated", options.Name)
			}
			time.Sleep(500 * time.Millisecond)
			nextGen, err = c.getServiceSpecGeneration(options.Namespace, options.Name)
			if err != nil {
				return err
			}
			if nextGen > gen {
				break
			}
		}
		if options.Verbose {
			go c.displayFunctionCreationProgress(ns, service.Name, log, stopChan, errChan)
		}
		err = c.waitForSuccessOrFailure(ns, service.Name, nextGen, stopChan, errChan)
		if err != nil {
			return err
		}
	}

	return nil
}

// publishImage returns true unless the name starts with 'dev.local' or 'ko.local'.
// These names have special meaning within knative and are the only Service
// images that will pull from the Docker deamon instead of a registry.
func publishImage(image string) bool {
	return strings.Index(image, "dev.local/") != 0 && strings.Index(image, "ko.local/") != 0
}
