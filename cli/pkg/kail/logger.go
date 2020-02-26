/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kail

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kail"
	"github.com/projectriff/cli/pkg/k8s"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/projectriff/system/pkg/apis/core/v1alpha1"
	knativev1alpha1 "github.com/projectriff/system/pkg/apis/knative/v1alpha1"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type Logger interface {
	ApplicationLogs(ctx context.Context, application *buildv1alpha1.Application, since time.Duration, out io.Writer) error
	FunctionLogs(ctx context.Context, function *buildv1alpha1.Function, since time.Duration, out io.Writer) error
	CoreDeployerLogs(ctx context.Context, deployer *corev1alpha1.Deployer, since time.Duration, out io.Writer) error
	StreamingProcessorLogs(ctx context.Context, processor *streamingv1alpha1.Processor, since time.Duration, out io.Writer) error
	KafkaGatewayLogs(ctx context.Context, gateway *streamingv1alpha1.KafkaGateway, since time.Duration, out io.Writer) error
	PulsarGatewayLogs(ctx context.Context, gateway *streamingv1alpha1.PulsarGateway, since time.Duration, out io.Writer) error
	InMemoryGatewayLogs(ctx context.Context, gateway *streamingv1alpha1.InMemoryGateway, since time.Duration, out io.Writer) error
	KnativeDeployerLogs(ctx context.Context, deployer *knativev1alpha1.Deployer, since time.Duration, out io.Writer) error
}

func NewDefault(k8s k8s.Client) Logger {
	return &logger{
		k8s: k8s,
	}
}

type logger struct {
	k8s k8s.Client
}

func (c *logger) ApplicationLogs(ctx context.Context, application *buildv1alpha1.Application, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", buildv1alpha1.ApplicationLabelKey, application.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return c.stream(ctx, application.Namespace, selector, containers, since, out)
}

func (c *logger) FunctionLogs(ctx context.Context, function *buildv1alpha1.Function, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", buildv1alpha1.FunctionLabelKey, function.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return c.stream(ctx, function.Namespace, selector, containers, since, out)
}

func (c *logger) CoreDeployerLogs(ctx context.Context, deployer *corev1alpha1.Deployer, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", corev1alpha1.DeployerLabelKey, deployer.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return c.stream(ctx, deployer.Namespace, selector, containers, since, out)
}

func (c *logger) StreamingProcessorLogs(ctx context.Context, processor *streamingv1alpha1.Processor, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", streamingv1alpha1.ProcessorLabelKey, processor.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{"function", "processor"}
	return c.stream(ctx, processor.Namespace, selector, containers, since, out)
}

func (c *logger) KafkaGatewayLogs(ctx context.Context, gateway *streamingv1alpha1.KafkaGateway, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", streamingv1alpha1.KafkaGatewayLabelKey, gateway.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return c.stream(ctx, gateway.Namespace, selector, containers, since, out)
}

func (c *logger) PulsarGatewayLogs(ctx context.Context, gateway *streamingv1alpha1.PulsarGateway, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", streamingv1alpha1.PulsarGatewayLabelKey, gateway.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return c.stream(ctx, gateway.Namespace, selector, containers, since, out)
}

func (c *logger) InMemoryGatewayLogs(ctx context.Context, gateway *streamingv1alpha1.InMemoryGateway, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", streamingv1alpha1.InMemoryGatewayLabelKey, gateway.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return c.stream(ctx, gateway.Namespace, selector, containers, since, out)
}

func (c *logger) KnativeDeployerLogs(ctx context.Context, deployer *knativev1alpha1.Deployer, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", knativev1alpha1.DeployerLabelKey, deployer.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{"user-container"}
	return c.stream(ctx, deployer.Namespace, selector, containers, since, out)
}

func (c *logger) stream(ctx context.Context, namespace string, selector labels.Selector, containers []string, since time.Duration, out io.Writer) error {
	// avoid kail logs appearing
	l := logutil.New(log.New(ioutil.Discard, "", log.LstdFlags), ioutil.Discard)
	ctx = logutil.NewContext(ctx, l)

	rc := c.k8s.KubeRestConfig()
	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return err
	}
	ds, err := kail.NewDSBuilder().WithNamespace(namespace).WithSelectors(selector).Create(ctx, cs)
	if err != nil {
		return err
	}
	filter := kail.NewContainerFilter(containers)
	controller, err := kail.NewController(ctx, cs, rc, ds.Pods(), filter, since)
	if err != nil {
		return err
	}
	writer := kail.NewWriter(out)
	for {
		select {
		case ev := <-controller.Events():
			writer.Print(ev)
		case <-controller.Done():
			return nil
		}
	}
}
