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

package commands_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/kballard/go-shellquote"
	"github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHandlerInvokeOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.HandlerInvokeOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldError: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.HandlerInvokeOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "json content type",
			Options: &commands.HandlerInvokeOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContentTypeJSON: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "text content type",
			Options: &commands.HandlerInvokeOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContentTypeText: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "multiple content types",
			Options: &commands.HandlerInvokeOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContentTypeJSON: true,
				ContentTypeText: true,
			},
			ExpectFieldError: cli.ErrMultipleOneOf(cli.JSONFlagName, cli.TextFlagName),
		},
	}

	table.Run(t)
}

func TestHandlerInvokeCommand(t *testing.T) {
	t.Parallel()

	handlerName := "test-handler"
	defaultNamespace := "default"

	handler := &requestv1alpha1.Handler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      handlerName,
		},
		Status: requestv1alpha1.HandlerStatus{
			Status: duckv1alpha1.Status{
				Conditions: []duckv1alpha1.Condition{
					{Type: requestv1alpha1.HandlerConditionReady, Status: "True"},
				},
			},
			URL: &apis.URL{
				Host: fmt.Sprintf("%s.example.com", handlerName),
			},
		},
	}

	ingressService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "istio-system",
			Name:      "istio-ingressgateway",
		},
		Spec: corev1.ServiceSpec{
			Type: "LoadBalancer",
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{Hostname: "localhost"},
				},
			},
		},
	}

	table := rifftesting.CommandTable{
		{
			Name:       "ingress loadbalancer hostname",
			Args:       []string{handlerName},
			ExecHelper: "HandlerInvoke",
			GivenObjects: []runtime.Object{
				handler,
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "istio-system",
						Name:      "istio-ingressgateway",
					},
					Spec: corev1.ServiceSpec{
						Type: "LoadBalancer",
					},
					Status: corev1.ServiceStatus{
						LoadBalancer: corev1.LoadBalancerStatus{
							Ingress: []corev1.LoadBalancerIngress{
								{Hostname: "localhost"},
							},
						},
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-handler.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "ingress loadbalancer ip",
			Args:       []string{handlerName},
			ExecHelper: "HandlerInvoke",
			GivenObjects: []runtime.Object{
				handler,
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "istio-system",
						Name:      "istio-ingressgateway",
					},
					Spec: corev1.ServiceSpec{
						Type: "LoadBalancer",
					},
					Status: corev1.ServiceStatus{
						LoadBalancer: corev1.LoadBalancerStatus{
							Ingress: []corev1.LoadBalancerIngress{
								{IP: "127.0.0.1"},
							},
						},
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl 127.0.0.1 -H 'Host: test-handler.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "ingress nodeport",
			Args:       []string{handlerName},
			ExecHelper: "HandlerInvoke",
			GivenObjects: []runtime.Object{
				handler,
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "istio-system",
						Name:      "istio-ingressgateway",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{Name: "http2", NodePort: 54321},
						},
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl http://localhost:54321 -H 'Host: test-handler.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "request path",
			Args:       []string{handlerName, "/path"},
			ExecHelper: "HandlerInvoke",
			GivenObjects: []runtime.Object{
				handler,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost/path -H 'Host: test-handler.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "content type json",
			Args:       []string{handlerName, cli.JSONFlagName},
			ExecHelper: "HandlerInvoke",
			GivenObjects: []runtime.Object{
				handler,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-handler.example.com' -H 'Content-Type: application/json'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "content type text",
			Args:       []string{handlerName, cli.TextFlagName},
			ExecHelper: "HandlerInvoke",
			GivenObjects: []runtime.Object{
				handler,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-handler.example.com' -H 'Content-Type: text/plain'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "pass extra args to curl",
			Args:       []string{handlerName, "--", "-w", "\n"},
			ExecHelper: "HandlerInvoke",
			GivenObjects: []runtime.Object{
				handler,
				ingressService,
			},
			ExpectOutput: `
Command executed: curl localhost -H 'Host: test-handler.example.com' -w '` + "\n" + `'
`,
		},
		{
			Name: "unknown ingress",
			Args: []string{handlerName},
			GivenObjects: []runtime.Object{
				handler,
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "istio-system",
						Name:      "istio-ingressgateway",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name:       "missing ingress",
			Args:       []string{handlerName},
			ExecHelper: "HandlerInvoke",
			GivenObjects: []runtime.Object{
				handler,
			},
			ShouldError: true,
		},
		{
			Name: "handler not ready",
			Args: []string{handlerName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      handlerName,
					},
					Status: requestv1alpha1.HandlerStatus{
						URL: &apis.URL{
							Host: fmt.Sprintf("%s.example.com", handlerName),
						},
					},
				},
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name: "handler missing domain",
			Args: []string{handlerName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      handlerName,
					},
					Status: requestv1alpha1.HandlerStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.HandlerConditionReady, Status: "True"},
							},
						},
					},
				},
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name: "missing handler",
			Args: []string{handlerName},
			GivenObjects: []runtime.Object{
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name:       "curl error",
			Args:       []string{handlerName},
			ExecHelper: "HandlerInvokeError",
			GivenObjects: []runtime.Object{
				handler,
				ingressService,
			},
			ExpectOutput: `
Command executed: curl localhost -H 'Host: test-handler.example.com'
`,
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewHandlerInvokeCommand)
}

func TestHelperProcess_HandlerInvoke(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stderr, "Command executed: %s\n", shellquote.Join(argsAfterBareDoubleDash(os.Args)...))
	os.Exit(0)
}

func TestHelperProcess_HandlerInvokeError(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stderr, "Command executed: %s\n", shellquote.Join(argsAfterBareDoubleDash(os.Args)...))
	os.Exit(1)
}

func argsAfterBareDoubleDash(args []string) []string {
	for i, arg := range args {
		if arg == "--" {
			return args[i+1:]
		}
	}
	return []string{}
}
