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

	"github.com/boz/go-logutil"
	"github.com/boz/kail"
	"github.com/projectriff/riff/pkg/k8s"
	"github.com/projectriff/system/pkg/apis/build"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/projectriff/system/pkg/apis/request"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	"github.com/projectriff/system/pkg/apis/stream"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/stream/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type Logger interface {
	ApplicationLogs(ctx context.Context, application *buildv1alpha1.Application, since time.Duration, out io.Writer) error
	FunctionLogs(ctx context.Context, function *buildv1alpha1.Function, since time.Duration, out io.Writer) error
	HandlerLogs(ctx context.Context, handler *requestv1alpha1.Handler, since time.Duration, out io.Writer) error
	ProcessorLogs(ctx context.Context, processor *streamv1alpha1.Processor, since time.Duration, out io.Writer) error
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
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", build.ApplicationLabelKey, application.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return c.stream(ctx, application.Namespace, selector, containers, since, out)
}

func (c *logger) FunctionLogs(ctx context.Context, function *buildv1alpha1.Function, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", build.FunctionLabelKey, function.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{}
	return c.stream(ctx, function.Namespace, selector, containers, since, out)
}

func (c *logger) HandlerLogs(ctx context.Context, handler *requestv1alpha1.Handler, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", request.HandlerLabelKey, handler.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{"user-container"}
	return c.stream(ctx, handler.Namespace, selector, containers, since, out)
}

func (c *logger) ProcessorLogs(ctx context.Context, processor *streamv1alpha1.Processor, since time.Duration, out io.Writer) error {
	selector, err := labels.Parse(fmt.Sprintf("%s=%s", stream.ProcessorLabelKey, processor.Name))
	if err != nil {
		panic(err)
	}
	containers := []string{"function", "processor"}
	return c.stream(ctx, processor.Namespace, selector, containers, since, out)
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
