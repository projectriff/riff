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

	"github.com/kballard/go-shellquote"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRequestProcessorInvokeOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.RequestProcessorInvokeOptions{
				ResourceOptions: testing.InvalidResourceOptions,
			},
			ExpectFieldError: testing.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.RequestProcessorInvokeOptions{
				ResourceOptions: testing.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "json content type",
			Options: &commands.RequestProcessorInvokeOptions{
				ResourceOptions: testing.ValidResourceOptions,
				ContentTypeJson: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "text content type",
			Options: &commands.RequestProcessorInvokeOptions{
				ResourceOptions: testing.ValidResourceOptions,
				ContentTypeText: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "multiple content types",
			Options: &commands.RequestProcessorInvokeOptions{
				ResourceOptions: testing.ValidResourceOptions,
				ContentTypeJson: true,
				ContentTypeText: true,
			},
			ExpectFieldError: cli.ErrMultipleOneOf("json", "text"),
		},
	}

	table.Run(t)
}

func TestRequestProcessorInvokeCommand(t *testing.T) {
	t.Parallel()

	requestprocessorName := "test-requestprocessor"
	defaultNamespace := "default"

	requestProcessor := &requestv1alpha1.RequestProcessor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: defaultNamespace,
			Name:      requestprocessorName,
		},
		Status: requestv1alpha1.RequestProcessorStatus{
			Status: duckv1alpha1.Status{
				Conditions: []duckv1alpha1.Condition{
					{Type: requestv1alpha1.RequestProcessorConditionReady, Status: "True"},
				},
			},
			Domain: fmt.Sprintf("%s.example.com", requestprocessorName),
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

	table := testing.CommandTable{
		{
			Name:       "ingress loadbalancer hostname",
			Args:       []string{requestprocessorName},
			ExecHelper: "RequestProcessorInvoke",
			GivenObjects: []runtime.Object{
				requestProcessor,
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
					"curl localhost -H 'Host: test-requestprocessor.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "ingress loadbalancer ip",
			Args:       []string{requestprocessorName},
			ExecHelper: "RequestProcessorInvoke",
			GivenObjects: []runtime.Object{
				requestProcessor,
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
					"curl 127.0.0.1 -H 'Host: test-requestprocessor.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "ingress nodeport",
			Args:       []string{requestprocessorName},
			ExecHelper: "RequestProcessorInvoke",
			GivenObjects: []runtime.Object{
				requestProcessor,
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
					"curl http://localhost:54321 -H 'Host: test-requestprocessor.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "request path",
			Args:       []string{requestprocessorName, "/path"},
			ExecHelper: "RequestProcessorInvoke",
			GivenObjects: []runtime.Object{
				requestProcessor,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost/path -H 'Host: test-requestprocessor.example.com'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "content type json",
			Args:       []string{requestprocessorName, "--json"},
			ExecHelper: "RequestProcessorInvoke",
			GivenObjects: []runtime.Object{
				requestProcessor,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-requestprocessor.example.com' -H 'Content-Type: application/json'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "content type text",
			Args:       []string{requestprocessorName, "--text"},
			ExecHelper: "RequestProcessorInvoke",
			GivenObjects: []runtime.Object{
				requestProcessor,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-requestprocessor.example.com' -H 'Content-Type: text/plain'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name:       "pass extra args to curl",
			Args:       []string{requestprocessorName, "--", "-w", "\n"},
			ExecHelper: "RequestProcessorInvoke",
			GivenObjects: []runtime.Object{
				requestProcessor,
				ingressService,
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-requestprocessor.example.com' -w '\n'\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name: "unknown ingress",
			Args: []string{requestprocessorName},
			GivenObjects: []runtime.Object{
				requestProcessor,
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
			Args:       []string{requestprocessorName},
			ExecHelper: "RequestProcessorInvoke",
			GivenObjects: []runtime.Object{
				requestProcessor,
			},
			ShouldError: true,
		},
		{
			Name: "request processor not ready",
			Args: []string{requestprocessorName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestprocessorName,
					},
					Status: requestv1alpha1.RequestProcessorStatus{
						Domain: fmt.Sprintf("%s.example.com", requestprocessorName),
					},
				},
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name: "request processor missing domain",
			Args: []string{requestprocessorName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestprocessorName,
					},
					Status: requestv1alpha1.RequestProcessorStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.RequestProcessorConditionReady, Status: "True"},
							},
						},
					},
				},
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name: "missing request processor",
			Args: []string{requestprocessorName},
			GivenObjects: []runtime.Object{
				ingressService,
			},
			ShouldError: true,
		},
		{
			Name:       "curl error",
			Args:       []string{requestprocessorName},
			ExecHelper: "RequestProcessorInvokeError",
			GivenObjects: []runtime.Object{
				requestProcessor,
				ingressService,
			},
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					"curl localhost -H 'Host: test-requestprocessor.example.com'\n",
					"exit status 255\n",
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
	}

	table.Run(t, commands.NewRequestProcessorInvokeCommand)
}

func TestHelperProcess_RequestProcessorInvoke(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// TODO assert arguments
	fmt.Fprintf(os.Stderr, "Command executed: %s\n", shellquote.Join(argsAfterBareDoubleDash(os.Args)...))
	os.Exit(0)
}

func TestHelperProcess_RequestProcessorInvokeError(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// TODO assert arguments
	fmt.Fprintf(os.Stderr, "Command executed: %s\n", shellquote.Join(argsAfterBareDoubleDash(os.Args)...))
	os.Exit(-1)
}

func argsAfterBareDoubleDash(args []string) []string {
	for i, arg := range args {
		if arg == "--" {
			return args[i+1:]
		}
	}
	return []string{}
}
