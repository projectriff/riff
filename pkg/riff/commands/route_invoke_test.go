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
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRouteInvokeOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.RouteInvokeOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldError: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.RouteInvokeOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "json content type",
			Options: &commands.RouteInvokeOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContentTypeJSON: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "text content type",
			Options: &commands.RouteInvokeOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContentTypeText: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "multiple content types",
			Options: &commands.RouteInvokeOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContentTypeJSON: true,
				ContentTypeText: true,
			},
			ExpectFieldError: cli.ErrMultipleOneOf(cli.JSONFlagName, cli.TextFlagName),
		},
	}

	table.Run(t)
}

func TestRouteInvokeCommand(t *testing.T) {
	t.Parallel()

	routeName := "test-route"
	defaultNamespace := "default"

	route := &requestv1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      routeName,
		},
		Status: requestv1alpha1.RouteStatus{
			Status: duckv1alpha1.Status{
				Conditions: []duckv1alpha1.Condition{
					{Type: requestv1alpha1.RouteConditionReady, Status: "True"},
				},
			},
			Domain: fmt.Sprintf("%s.example.com", routeName),
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
			Args:       []string{routeName},
			ExecHelper: "RouteInvoke",
			GivenObjects: []runtime.Object{
				route,
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
					"curl localhost -H 'Host: test-route.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "ingress loadbalancer ip",
			Args:       []string{routeName},
			ExecHelper: "RouteInvoke",
			GivenObjects: []runtime.Object{
				route,
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
					"curl 127.0.0.1 -H 'Host: test-route.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "ingress nodeport",
			Args:       []string{routeName},
			ExecHelper: "RouteInvoke",
			GivenObjects: []runtime.Object{
				route,
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
					"curl http://localhost:54321 -H 'Host: test-route.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "request path",
			Args:       []string{routeName, "/path"},
			ExecHelper: "RouteInvoke",
			GivenObjects: []runtime.Object{
				route,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost/path -H 'Host: test-route.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "content type json",
			Args:       []string{routeName, cli.JSONFlagName},
			ExecHelper: "RouteInvoke",
			GivenObjects: []runtime.Object{
				route,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-route.example.com' -H 'Content-Type: application/json'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "content type text",
			Args:       []string{routeName, cli.TextFlagName},
			ExecHelper: "RouteInvoke",
			GivenObjects: []runtime.Object{
				route,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-route.example.com' -H 'Content-Type: text/plain'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "pass extra args to curl",
			Args:       []string{routeName, "--", "-w", "\n"},
			ExecHelper: "RouteInvoke",
			GivenObjects: []runtime.Object{
				route,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-route.example.com' -w '\n'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name: "unknown ingress",
			Args: []string{routeName},
			GivenObjects: []runtime.Object{
				route,
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
			Args:       []string{routeName},
			ExecHelper: "RouteInvoke",
			GivenObjects: []runtime.Object{
				route,
			},
			ShouldError: true,
		},
		{
			Name: "route not ready",
			Args: []string{routeName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
					Status: requestv1alpha1.RouteStatus{
						Domain: fmt.Sprintf("%s.example.com", routeName),
					},
				},
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name: "route missing domain",
			Args: []string{routeName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
					Status: requestv1alpha1.RouteStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.RouteConditionReady, Status: "True"},
							},
						},
					},
				},
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name: "missing route",
			Args: []string{routeName},
			GivenObjects: []runtime.Object{
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name:       "curl error",
			Args:       []string{routeName},
			ExecHelper: "RouteInvokeError",
			GivenObjects: []runtime.Object{
				route,
				ingressService,
			},
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-route.example.com'\n",
					"exit status 1\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
	}

	table.Run(t, commands.NewRouteInvokeCommand)
}

func TestHelperProcess_RouteInvoke(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stderr, "Command executed: %s\n", shellquote.Join(argsAfterBareDoubleDash(os.Args)...))
	os.Exit(0)
}

func TestHelperProcess_RouteInvokeError(t *testing.T) {
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
