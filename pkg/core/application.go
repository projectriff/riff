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
	errs "errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/boz/kail"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	projectriff "github.com/projectriff/system/pkg/apis/projectriff/v1alpha1"
	projectriffv1alpha1 "github.com/projectriff/system/pkg/apis/projectriff/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	applicationLabel              = "riff.projectriff.io/application"
	applicationArtifactAnnotation = "riff.projectriff.io/artifact"
	applicationOverrideAnnotation = "riff.projectriff.io/override"
	applicationHandlerAnnotation  = "riff.projectriff.io/handler"
)

type CreateApplicationOptions struct {
	CreateOrUpdateServiceOptions
	BuildOptions

	GitRepo     string
	GitRevision string
	SubPath     string
}

type ListApplicationOptions struct {
	Namespace string
}

func (c *client) ListApplications(options ListApplicationOptions) (*projectriffv1alpha1.ApplicationList, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.system.ProjectriffV1alpha1().Applications(ns).List(meta_v1.ListOptions{})
}

func (c *client) CreateApplication(buildpackBuilder Builder, options CreateApplicationOptions, log io.Writer) (*projectriffv1alpha1.Application, error) {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	applicationName := options.Name
	_, err := c.system.ProjectriffV1alpha1().Applications(ns).Get(applicationName, v1.GetOptions{})
	if err == nil {
		return nil, fmt.Errorf("application '%s' already exists in namespace '%s'", applicationName, ns)
	}

	f, err := newApplication(options)
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
		f.Spec.Build.Template = "cnb"
		f.Spec.Build.Source = c.makeApplicationBuildSourceSpec(options)
	}

	if !options.DryRun {
		f, err := c.system.ProjectriffV1alpha1().Applications(ns).Create(f)
		if err != nil {
			return nil, err
		}

		if options.Verbose || options.Wait {
			stopChan := make(chan struct{})
			errChan := make(chan error)
			if options.Verbose {
				go c.displayApplicationCreationProgress(ns, f.Name, log, stopChan, errChan)
			}
			err := c.waitForSuccessOrFailure(ns, f.Name, 1, stopChan, errChan, options.Verbose)
			if err != nil {
				return nil, err
			}
		}
	}

	return f, nil
}

func newApplication(options CreateApplicationOptions) (*projectriffv1alpha1.Application, error) {
	envVars, err := ParseEnvVar(options.Env)
	if err != nil {
		return nil, err
	}
	envVarsFrom, err := ParseEnvVarSource(options.EnvFrom)
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, envVarsFrom...)

	f := projectriffv1alpha1.Application{
		TypeMeta: meta_v1.TypeMeta{
			APIVersion: "projectriff.io/v1alpha1",
			Kind:       "Application",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name: options.Name,
		},
		Spec: projectriffv1alpha1.ApplicationSpec{
			Image: options.Image,
			Build: projectriffv1alpha1.ApplicationBuild{},
			Run: projectriffv1alpha1.ApplicationRun{
				Env: envVars,
			},
		},
	}

	return &f, nil
}

func (c *client) makeApplicationBuildSourceSpec(options CreateApplicationOptions) *projectriffv1alpha1.Source {
	return &projectriffv1alpha1.Source{
		Git: &projectriffv1alpha1.GitSource{
			URL:      options.GitRepo,
			Revision: options.GitRevision,
		},
		SubPath: options.SubPath,
	}
}

func (c *client) displayApplicationCreationProgress(applicationNamespace string, applicationName string, logWriter io.Writer, stopChan <-chan struct{}, errChan chan<- error) {
	// TODO this assumes the application and service have the same name, they may not in the future
	revName, err := c.revisionName(applicationNamespace, applicationName, logWriter, stopChan)
	if err != nil {
		errChan <- err
		return
	} else if revName == "" { // stopped
		return
	}
	buildName, err := c.buildName(applicationNamespace, revName, logWriter, stopChan)
	if err != nil {
		errChan <- err
		return
	} else if buildName == "" { // stopped
		return
	}

	ctx := newContext()

	podController, err := c.podController(buildName, applicationName, ctx)
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

type UpdateApplicationOptions struct {
	Namespace string
	Name      string
	LocalPath string
	Verbose   bool
	Wait      bool
}

func (c *client) getApplicationSpecGeneration(namespace string, name string) (int64, error) {
	f, err := c.application(namespace, name)
	if err != nil {
		return 0, err
	}
	return f.Generation, nil
}

func (c *client) UpdateApplication(buildpackBuilder Builder, options UpdateApplicationOptions, log io.Writer) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)

	application, err := c.application(ns, options.Name)
	if err != nil {
		return err
	}

	// create a copy before mutating
	application = application.DeepCopy()

	gen := application.Generation

	appDir := options.LocalPath

	if appDir != "" {
		// application was built locally, attempt to reconstruct configuration
		localBuild := BuildOptions{
			LocalPath: appDir,
		}
		repoName := application.Spec.Image

		err := c.doBuildLocally(buildpackBuilder, repoName, localBuild, log)
		if err != nil {
			return err
		}
		// TODO this won't work, we need a replacement
		// c.bumpBuildAnnotationForRevision(service)
	} else {
		if application.Spec.Build.Source == nil {
			return fmt.Errorf("local-path must be specified to rebuild application from source")
		}
		// TODO this won't work, we need a replacement
		// c.bumpBuildAnnotationForBuild(service)
	}

	_, err = c.system.ProjectriffV1alpha1().Applications(application.Namespace).Update(application)
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
				return fmt.Errorf("update unsuccessful for \"%s\", application resource was never updated", options.Name)
			}
			time.Sleep(500 * time.Millisecond)
			nextGen, err = c.getApplicationSpecGeneration(ns, options.Name)
			if err != nil {
				return err
			}
			if nextGen > gen {
				break
			}
		}
		if options.Verbose {
			go c.displayApplicationCreationProgress(ns, application.Name, log, stopChan, errChan)
		}
		err = c.waitForSuccessOrFailure(ns, application.Name, nextGen, stopChan, errChan, options.Verbose)
		if err != nil {
			return err
		}
	}

	return nil
}

type BuildApplicationOptions struct {
	BuildOptions

	Image string
}

func (c *client) BuildApplication(buildpackBuilder Builder, options BuildApplicationOptions, log io.Writer) error {
	return c.doBuildLocally(buildpackBuilder, options.Image, options.BuildOptions, log)
}

type ApplicationStatusOptions struct {
	Namespace string
	Name      string
}

func (c *client) ApplicationStatus(options ApplicationStatusOptions) (*duckv1alpha1.Condition, error) {

	s, err := c.application(options.Namespace, options.Name)
	if err != nil {
		return nil, err
	}

	if ready := s.Status.GetCondition(projectriffv1alpha1.ApplicationConditionReady); ready != nil {
		return ready, nil
	}
	return nil, errs.New("No condition of type ApplicationConditionReady found for the application")
}

type ApplicationInvokeOptions struct {
	Namespace       string
	Name            string
	ContentTypeText bool
	ContentTypeJson bool
}

func (c *client) ApplicationCoordinates(options ApplicationInvokeOptions) (string, string, error) {

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

	f, err := c.application(options.Namespace, options.Name)
	if err != nil {
		return "", "", err
	}

	return ingress, f.Status.Address.Hostname, nil
}

func (c *client) application(namespace, name string) (*projectriff.Application, error) {
	ns := c.explicitOrConfigNamespace(namespace)
	return c.system.ProjectriffV1alpha1().Applications(ns).Get(name, meta_v1.GetOptions{})
}

type DeleteApplicationOptions struct {
	Namespace string
	Name      string
}

func (c *client) DeleteApplication(options DeleteApplicationOptions) error {
	ns := c.explicitOrConfigNamespace(options.Namespace)
	return c.system.ProjectriffV1alpha1().Applications(ns).Delete(options.Name, nil)
}
