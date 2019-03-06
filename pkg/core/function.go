/*
 * Copyright 2018-2019 The original author or authors
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
	"os"
	"path/filepath"
	"time"

	logutil "github.com/boz/go-logutil"

	"github.com/BurntSushi/toml"

	"github.com/boz/kail"
	"github.com/boz/kcache/types/pod"
	build "github.com/knative/build/pkg/apis/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/projectriff/riff/pkg/env"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	functionLabel              = "riff.projectriff.io/function"
	buildAnnotation            = "riff.projectriff.io/build"
	functionArtifactAnnotation = "riff.projectriff.io/artifact"
	functionOverrideAnnotation = "riff.projectriff.io/override"
	functionHandlerAnnotation  = "riff.projectriff.io/handler"
	buildTemplateName          = "riff-cnb"
	pollServiceTimeout         = 10 * time.Minute
	pollServicePollingInterval = time.Second
)

type BuildOptions struct {
	Invoker        string
	Handler        string
	Artifact       string
	LocalPath      string
	BuildpackImage string
	RunImage       string
}
type CreateFunctionOptions struct {
	CreateOrUpdateServiceOptions
	BuildOptions

	GitRepo     string
	GitRevision string
}

func (c *client) CreateFunction(buildpackBuilder Builder, options CreateFunctionOptions, log io.Writer) (*v1alpha1.Service, *corev1.PersistentVolumeClaim, error) {
	var buildCache *corev1.PersistentVolumeClaim
	ns := c.explicitOrConfigNamespace(options.Namespace)
	functionName := options.Name
	_, err := c.serving.ServingV1alpha1().Services(ns).Get(functionName, v1.GetOptions{})
	if err == nil {
		return nil, nil, fmt.Errorf("service '%s' already exists in namespace '%s'", functionName, ns)
	}

	s, err := newService(options.CreateOrUpdateServiceOptions)
	if err != nil {
		return nil, nil, err
	}

	labels := s.Spec.RunLatest.Configuration.RevisionTemplate.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labels[functionLabel] = functionName
	s.Spec.RunLatest.Configuration.RevisionTemplate.SetLabels(labels)

	if options.LocalPath != "" {
		if s.ObjectMeta.Annotations == nil {
			s.ObjectMeta.Annotations = make(map[string]string)
		}
		s.ObjectMeta.Annotations[functionArtifactAnnotation] = options.Artifact
		s.ObjectMeta.Annotations[functionHandlerAnnotation] = options.Handler
		s.ObjectMeta.Annotations[functionOverrideAnnotation] = options.Invoker

		if options.DryRun {
			// skip build for a dry run
			log.Write([]byte("Skipping local build\n"))
		} else {
			if err := c.doBuildLocally(buildpackBuilder, options.Image, options.BuildOptions); err != nil {
				return nil, nil, err
			}
		}
		c.bumpBuildAnnotationForRevision(s)
	} else {
		// create build cache for service
		buildCache = &corev1.PersistentVolumeClaim{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1",
				Kind:       "PersistentVolumeClaim",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: fmt.Sprintf("%s-build-cache", options.Name),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("8Gi"),
					},
				},
			},
		}

		// buildpack based cluster build
		s.Spec.RunLatest.Configuration.Build = &v1alpha1.RawExtension{
			Object: &build.Build{
				TypeMeta: v1.TypeMeta{
					APIVersion: "build.knative.dev/v1alpha1",
					Kind:       "Build",
				},
				Spec: build.BuildSpec{
					ServiceAccountName: "riff-build",
					Source:             c.makeBuildSourceSpec(options),
					Template: &build.TemplateInstantiationSpec{
						Name: buildTemplateName,
						Kind: "ClusterBuildTemplate",
						Arguments: []build.ArgumentSpec{
							{Name: "IMAGE", Value: options.Image},
							{Name: "FUNCTION_ARTIFACT", Value: options.Artifact},
							{Name: "FUNCTION_HANDLER", Value: options.Handler},
							{Name: "FUNCTION_LANGUAGE", Value: options.Invoker},
							{Name: "CACHE_VOLUME", Value: buildCache.Name},
						},
					},
				},
			},
		}
		c.bumpBuildAnnotationForBuild(s)
	}

	if !options.DryRun {
		if buildCache != nil {
			buildCache, err = c.kubeClient.CoreV1().PersistentVolumeClaims(ns).Create(buildCache)
			if err != nil {
				return nil, nil, err
			}
		}

		s, err := c.serving.ServingV1alpha1().Services(ns).Create(s)
		if err != nil {
			return nil, nil, err
		}

		if buildCache != nil {
			// add service as owner of build cache so that deleting the service, deletes the cache
			if _, err = c.addOwnerReferenceToCacheWithRetry(3, buildCache, s); err != nil {
				return nil, nil, err
			}
		}

		if options.Verbose || options.Wait {
			stopChan := make(chan struct{})
			errChan := make(chan error)
			if options.Verbose {
				go c.displayFunctionCreationProgress(ns, s.Name, log, stopChan, errChan)
			}
			err := c.waitForSuccessOrFailure(ns, s.Name, 1, stopChan, errChan, options.Verbose)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return s, buildCache, nil
}

func (c *client) addOwnerReferenceToCacheWithRetry(attempts int, cache *corev1.PersistentVolumeClaim, svc *v1alpha1.Service) (*corev1.PersistentVolumeClaim, error) {
	result, err := c.addOwnerReferenceToCache(cache, svc)
	if err != nil && attempts > 0 {
		return c.addOwnerReferenceToCacheWithRetry(attempts-1, cache, svc)
	}
	return result, err
}

func (c *client) addOwnerReferenceToCache(cache *corev1.PersistentVolumeClaim, svc *v1alpha1.Service) (*corev1.PersistentVolumeClaim, error) {
	cache, err := c.kubeClient.CoreV1().PersistentVolumeClaims(cache.Namespace).Get(cache.Name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	cache.ObjectMeta.OwnerReferences = []v1.OwnerReference{
		*v1.NewControllerRef(svc, schema.GroupVersionKind{
			Group:   v1alpha1.SchemeGroupVersion.Group,
			Version: v1alpha1.SchemeGroupVersion.Version,
			Kind:    "Service",
		}),
	}
	return c.kubeClient.CoreV1().PersistentVolumeClaims(cache.Namespace).Update(cache)
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
	buildName, err := c.buildName(serviceNamespace, revName, logWriter, stopChan)
	if err != nil {
		errChan <- err
		return
	} else if buildName == "" { // stopped
		return
	}

	ctx := newContext()

	podController, err := c.podController(buildName, serviceName, ctx)
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
	fmt.Fprintf(logWriter, "Waiting for LatestCreatedRevisionName\n")
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
		select {
		case <-stopChan:
			return "", nil
		default:
			// continue
		}
	}
	fmt.Fprintf(logWriter, "LatestCreatedRevisionName available: %s\n", revName)
	return revName, nil
}

func (c *client) buildName(ns string, revName string, logWriter io.Writer, stopChan <-chan struct{}) (string, error) {
	revObj, err := c.serving.ServingV1alpha1().Revisions(ns).Get(revName, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	if revObj.Spec.BuildRef == nil {
		// revsion has no build
		return "", nil
	}
	return revObj.Spec.BuildRef.Name, nil
}

func newContext() context.Context {
	ctx := context.Background()
	// avoid kail logs appearing
	l := logutil.New(log.New(ioutil.Discard, "", log.LstdFlags), ioutil.Discard)
	ctx = logutil.NewContext(ctx, l)
	return ctx
}

func (c *client) podController(buildName string, serviceName string, ctx context.Context) (pod.Controller, error) {
	dsb := kail.NewDSBuilder()

	buildSel, err := labels.Parse(fmt.Sprintf("%s=%s", "build.knative.dev/buildName", buildName))
	if err != nil {
		return nil, err
	}
	runtimeSel, err := labels.Parse(fmt.Sprintf("%s=%s", "serving.knative.dev/configuration", serviceName))
	if err != nil {
		return nil, err
	}
	ds, err := dsb.WithSelectors(or(buildSel, runtimeSel)).Create(ctx, c.kubeClient)
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

type serviceChecker func() (transientErr error, err error)

func (c *client) createServiceChecker(namespace string, name string, gen int64) serviceChecker {
	return func() (transientErr error, err error) {
		return checkService(c, namespace, name, gen)
	}
}

func (c *client) waitForSuccessOrFailure(namespace string, name string, gen int64, stopChan chan<- struct{}, errChan <-chan error, verbose bool) error {
	defer close(stopChan)

	var log io.Writer
	if verbose {
		log = os.Stdout
	} else {
		log = ioutil.Discard
	}

	check := c.createServiceChecker(namespace, name, gen)

	return pollService(check, errChan, pollServiceTimeout, pollServicePollingInterval, log)
}

func pollService(check serviceChecker, errChan <-chan error, timeout time.Duration, sleepDuration time.Duration, log io.Writer) error {
	sleepTime := time.Duration(0)
	lastTransientErr := ""
	for {
		select {
		case err := <-errChan:
			return err
		default:
		}

		transientError, err := check()
		if err != nil {
			return err
		}

		if transientError == nil {
			return nil
		}

		if sleepTime >= timeout {
			fmt.Fprintln(log, "Waiting on function creation timed out")
			return transientError
		}

		if te := transientError.Error(); te != lastTransientErr {
			fmt.Fprintf(log, "Waiting on function creation: %v\n", transientError)
			lastTransientErr = te
		}

		time.Sleep(sleepDuration)
		sleepTime += sleepDuration
	}
	return nil
}

func checkService(c *client, namespace string, name string, gen int64) (transientErr error, err error) {
	// TODO: Test this
	service, err := c.service(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("checkService failed to obtain service: %v", err)
	}

	if service.Status.ObservedGeneration < gen {
		// allow some time for service status observed generation to show up
		return fmt.Errorf("checkService failed to obtain service status for observedGeneration %d", gen), nil
	}

	if service.Status.IsReady() {
		return nil, nil
	}

	ready := service.Status.GetCondition(v1alpha1.ServiceConditionReady)
	if ready == nil {
		return nil, fmt.Errorf("unable to obtain ready condition status")
	}

	if ready.Status == corev1.ConditionFalse {
		if s := fetchTransientError(service.Status.Conditions); s != "" {
			return fmt.Errorf("%s: %s", s, ready.Reason), nil
		}
		return nil, fmt.Errorf("function creation failed: %s", ready.Reason)
	}
	return fmt.Errorf("function creation incomplete: service status unknown: %s", ready.Reason), nil
}

func fetchTransientError(conds duckv1alpha1.Conditions) string {
	for _, c := range conds {
		if c.IsUnknown() {
			return "function creation incomplete: service status false"
		}
	}
	return ""
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

type UpdateFunctionOptions struct {
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
	return s.Generation, nil
}

func (c *client) UpdateFunction(buildpackBuilder Builder, options UpdateFunctionOptions, log io.Writer) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	service, err := c.service(ns, options.Name)
	if err != nil {
		return err
	}

	// create a copy before mutating
	service = service.DeepCopy()

	gen := service.Generation

	// TODO support non-RunLatest configurations
	configuration := service.Spec.RunLatest.Configuration
	labels := configuration.RevisionTemplate.Labels
	if labels[functionLabel] == "" {
		return fmt.Errorf("the service named \"%s\" is not a %s function", options.Name, env.Cli.Name)
	}

	build := configuration.Build
	annotations := service.Annotations

	appDir := options.LocalPath
	if build != nil && appDir != "" {
		return fmt.Errorf("unable to proceed: local path specified for cluster-built service named \"%s\"", options.Name)
	}

	if build == nil {
		// function was built locally, attempt to reconstruct configuration
		localBuild := BuildOptions{
			LocalPath: appDir,
			Artifact:  annotations[functionArtifactAnnotation],
			Handler:   annotations[functionHandlerAnnotation],
			Invoker:   annotations[functionOverrideAnnotation],
		}
		repoName := configuration.RevisionTemplate.Spec.Container.Image
		if appDir == "" {
			return fmt.Errorf("local-path must be specified to rebuild function from source")
		}

		err := c.doBuildLocally(buildpackBuilder, repoName, localBuild)
		if err != nil {
			return err
		}
		c.bumpBuildAnnotationForRevision(service)
	} else {
		c.bumpBuildAnnotationForBuild(service)
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
				return fmt.Errorf("update unsuccessful for \"%s\", service resource was never updated", options.Name)
			}
			time.Sleep(500 * time.Millisecond)
			nextGen, err = c.getServiceSpecGeneration(ns, options.Name)
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
		err = c.waitForSuccessOrFailure(ns, service.Name, nextGen, stopChan, errChan, options.Verbose)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) doBuildLocally(builder Builder, image string, options BuildOptions) error {
	if err := writeRiffToml(options); err != nil {
		return err
	}
	defer func() { _ = deleteRiffToml(options) }()
	if options.BuildpackImage == "" || options.RunImage == "" {
		config, err := c.FetchPackConfig()
		if err != nil {
			return fmt.Errorf("unable to load pack config: %s", err)
		}
		options.BuildpackImage = config.BuilderImage
		options.RunImage = config.RunImage
	}
	return builder.Build(options.LocalPath, options.BuildpackImage, options.RunImage, image)
}

func writeRiffToml(options BuildOptions) error {
	t := struct {
		Override string `toml:"override"`
		Handler  string `toml:"handler"`
		Artifact string `toml:"artifact"`
	}{
		Override: options.Invoker,
		Handler:  options.Handler,
		Artifact: options.Artifact,
	}
	path := filepath.Join(options.LocalPath, "riff.toml")
	if _, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		return fmt.Errorf("found riff.toml file in local path. Please delete this file and let the CLI create it from flags")
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(t)
}

func deleteRiffToml(options BuildOptions) error {
	path := filepath.Join(options.LocalPath, "riff.toml")
	return os.Remove(path)
}
