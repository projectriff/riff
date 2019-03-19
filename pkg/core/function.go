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
	errs "errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	logutil "github.com/boz/go-logutil"

	"github.com/BurntSushi/toml"

	"github.com/boz/kail"
	"github.com/boz/kcache/types/pod"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	projectriff "github.com/projectriff/system/pkg/apis/projectriff/v1alpha1"
	projectriffv1alpha1 "github.com/projectriff/system/pkg/apis/projectriff/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	SubPath     string
}

type ListFunctionOptions struct {
	Namespace string
}

func (c *client) ListFunctions(options ListFunctionOptions) (*projectriffv1alpha1.FunctionList, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.system.ProjectriffV1alpha1().Functions(ns).List(meta_v1.ListOptions{})
}

func (c *client) CreateFunction(buildpackBuilder Builder, options CreateFunctionOptions, log io.Writer) (*projectriffv1alpha1.Function, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	functionName := options.Name
	_, err := c.system.ProjectriffV1alpha1().Functions(ns).Get(functionName, v1.GetOptions{})
	if err == nil {
		return nil, fmt.Errorf("function '%s' already exists in namespace '%s'", functionName, ns)
	}

	f, err := newFunction(options)
	if err != nil {
		return nil, err
	}

	if options.LocalPath != "" {
		if options.DryRun {
			// skip build for a dry run
			log.Write([]byte("Skipping local build\n"))
		} else {
			if err := c.doBuildLocally(buildpackBuilder, options.Image, options.BuildOptions, log); err != nil {
				return nil, err
			}
		}
	} else {
		// buildpack based cluster build
		cacheSize := resource.MustParse("8Gi")
		f.Spec.Build.CacheSize = &cacheSize
		f.Spec.Build.Source = c.makeBuildSourceSpec(options)
	}

	if !options.DryRun {
		f, err := c.system.ProjectriffV1alpha1().Functions(ns).Create(f)
		if err != nil {
			return nil, err
		}

		if options.Verbose || options.Wait {
			stopChan := make(chan struct{})
			errChan := make(chan error)
			if options.Verbose {
				go c.displayFunctionCreationProgress(ns, f.Name, log, stopChan, errChan)
			}
			err := c.waitForSuccessOrFailure(ns, f.Name, 1, stopChan, errChan, options.Verbose)
			if err != nil {
				return nil, err
			}
		}
	}

	return f, nil
}

func newFunction(options CreateFunctionOptions) (*projectriffv1alpha1.Function, error) {
	envVars, err := ParseEnvVar(options.Env)
	if err != nil {
		return nil, err
	}
	envVarsFrom, err := ParseEnvVarSource(options.EnvFrom)
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, envVarsFrom...)

	f := projectriffv1alpha1.Function{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: "projectriff.io/v1alpha1",
			Kind:       "Function",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: options.Name,
		},
		Spec: projectriffv1alpha1.FunctionSpec{
			Image: options.Image,
			Build: projectriffv1alpha1.FunctionBuild{
				Artifact: options.Artifact,
				Handler:  options.Handler,
				Invoker:  options.Invoker,
			},
			Run: projectriffv1alpha1.FunctionRun{
				Env: envVars,
			},
		},
	}

	return &f, nil
}

func (c *client) makeBuildSourceSpec(options CreateFunctionOptions) *projectriffv1alpha1.Source {
	return &projectriffv1alpha1.Source{
		Git: &projectriffv1alpha1.GitSource{
			URL:      options.GitRepo,
			Revision: options.GitRevision,
		},
		SubPath: options.SubPath,
	}
}

func (c *client) displayFunctionCreationProgress(functionNamespace string, functionName string, logWriter io.Writer, stopChan <-chan struct{}, errChan chan<- error) {
	// TODO this assumes the function and service have the same name, they may not in the future
	revName, err := c.revisionName(functionNamespace, functionName, logWriter, stopChan)
	if err != nil {
		errChan <- err
		return
	} else if revName == "" { // stopped
		return
	}
	buildName, err := c.buildName(functionNamespace, revName, logWriter, stopChan)
	if err != nil {
		errChan <- err
		return
	} else if buildName == "" { // stopped
		return
	}

	ctx := newContext()

	podController, err := c.podController(buildName, functionName, ctx)
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
			if errors.IsNotFound(err) {
				continue
			}
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

	// give a moment for resource to settle
	time.Sleep(5000 * time.Millisecond)

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

func (c *client) getFunctionSpecGeneration(namespace string, name string) (int64, error) {
	f, err := c.function(namespace, name)
	if err != nil {
		return 0, err
	}
	return f.Generation, nil
}

func (c *client) UpdateFunction(buildpackBuilder Builder, options UpdateFunctionOptions, log io.Writer) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	function, err := c.function(ns, options.Name)
	if err != nil {
		return err
	}

	// create a copy before mutating
	function = function.DeepCopy()

	gen := function.Generation

	appDir := options.LocalPath

	if appDir != "" {
		// function was built locally, attempt to reconstruct configuration
		localBuild := BuildOptions{
			LocalPath: appDir,
			Artifact:  function.Spec.Build.Artifact,
			Handler:   function.Spec.Build.Handler,
			Invoker:   function.Spec.Build.Invoker,
		}
		repoName := function.Spec.Image

		err := c.doBuildLocally(buildpackBuilder, repoName, localBuild, log)
		if err != nil {
			return err
		}
		// TODO this won't work, we need a replacement
		// c.bumpBuildAnnotationForRevision(service)
	} else {
		if function.Spec.Build.Source == nil {
			return fmt.Errorf("local-path must be specified to rebuild function from source")
		}
		// TODO this won't work, we need a replacement
		// c.bumpBuildAnnotationForBuild(service)
	}

	_, err = c.system.ProjectriffV1alpha1().Functions(function.Namespace).Update(function)
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
				return fmt.Errorf("update unsuccessful for \"%s\", function resource was never updated", options.Name)
			}
			time.Sleep(500 * time.Millisecond)
			nextGen, err = c.getFunctionSpecGeneration(ns, options.Name)
			if err != nil {
				return err
			}
			if nextGen > gen {
				break
			}
		}
		if options.Verbose {
			go c.displayFunctionCreationProgress(ns, function.Name, log, stopChan, errChan)
		}
		err = c.waitForSuccessOrFailure(ns, function.Name, nextGen, stopChan, errChan, options.Verbose)
		if err != nil {
			return err
		}
	}

	return nil
}

type BuildFunctionOptions struct {
	BuildOptions

	Image string
}

func (c *client) BuildFunction(buildpackBuilder Builder, options BuildFunctionOptions, log io.Writer) error {
	return c.doBuildLocally(buildpackBuilder, options.Image, options.BuildOptions, log)
}

func (c *client) doBuildLocally(builder Builder, image string, options BuildOptions, log io.Writer) error {
	return doLocally(options, func() error {
		if options.BuildpackImage == "" || options.RunImage == "" {
			config, err := c.FetchPackConfig()
			if err != nil {
				return fmt.Errorf("unable to load pack config: %s", err)
			}
			options.BuildpackImage = config.BuilderImage
			options.RunImage = config.RunImage
		}
		return builder.Build(options.LocalPath, options.BuildpackImage, options.RunImage, image, log)
	})
}

func doLocally(options BuildOptions, doer func() error) error {
	if err := writeRiffToml(options); err != nil {
		return err
	}
	defer func() { _ = deleteRiffToml(options) }()
	return doer()
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

type FunctionStatusOptions struct {
	Namespace string
	Name      string
}

func (c *client) FunctionStatus(options FunctionStatusOptions) (*duckv1alpha1.Condition, error) {

	s, err := c.function(options.Namespace, options.Name)
	if err != nil {
		return nil, err
	}

	if ready := s.Status.GetCondition(projectriffv1alpha1.FunctionConditionReady); ready != nil {
		return ready, nil
	}
	return nil, errs.New("No condition of type FunctionConditionReady found for the function")
}

type FunctionInvokeOptions struct {
	Namespace       string
	Name            string
	ContentTypeText bool
	ContentTypeJson bool
}

func (c *client) FunctionCoordinates(options FunctionInvokeOptions) (string, string, error) {

	ksvc, err := c.kubeClient.CoreV1().Services(istioNamespace).Get(ingressServiceName, meta_v1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	var ingress string
	if ksvc.Spec.Type == "LoadBalancer" {
		ingresses := ksvc.Status.LoadBalancer.Ingress
		if len(ingresses) > 0 {
			ingress = ingresses[0].IP
			if ingress == "" {
				ingress = ingresses[0].Hostname
			}
		}
	}
	if ingress == "" {
		for _, port := range ksvc.Spec.Ports {
			if port.Name == "http" || port.Name == "http2" {
				config, err := c.clientConfig.ClientConfig()
				if err != nil {
					return "", "", err
				}
				host := config.Host[0:strings.LastIndex(config.Host, ":")]
				host = strings.Replace(host, "https", "http", 1)
				ingress = fmt.Sprintf("%s:%d", host, port.NodePort)
			}
		}
		if ingress == "" {
			return "", "", errs.New("Ingress not available")
		}
	}

	f, err := c.function(options.Namespace, options.Name)
	if err != nil {
		return "", "", err
	}

	return ingress, f.Status.Address.Hostname, nil
}

func (c *client) function(namespace, name string) (*projectriff.Function, error) {
	ns := c.explicitOrConfigNamespace(namespace)
	return c.system.ProjectriffV1alpha1().Functions(ns).Get(name, meta_v1.GetOptions{})
}

type DeleteFunctionOptions struct {
	Namespace string
	Name      string
}

func (c *client) DeleteFunction(options DeleteFunctionOptions) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.system.ProjectriffV1alpha1().Functions(ns).Delete(options.Name, nil)
}
