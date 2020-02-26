/*
Copyright 2020 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package streaming_test

import (
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	streamingv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/controllers/streaming"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/riff/system/pkg/tracker"
)

func TestGatewayReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testName := "test-gateway"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}

	ports := []corev1.ServicePort{
		corev1.ServicePort{Name: "gateway", Port: 6565},
		corev1.ServicePort{Name: "provisioner", Port: 80, TargetPort: intstr.FromInt(8080)},
	}

	gatewayConditionDeploymentReady := factories.Condition().Type(streamingv1alpha1.GatewayConditionDeploymentReady)
	gatewayConditionReady := factories.Condition().Type(streamingv1alpha1.GatewayConditionReady)
	gatewayConditionServiceReady := factories.Condition().Type(streamingv1alpha1.GatewayConditionServiceReady)
	deploymentConditionAvailable := factories.Condition().Type("Available")
	deploymentConditionProgressing := factories.Condition().Type("Progressing")

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = streamingv1alpha1.AddToScheme(scheme)

	gatewayMinimal := factories.Gateway().
		NamespaceName(testNamespace, testName).
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.Generation(1)
		})
	gateway := gatewayMinimal.
		Ports(ports...).
		StatusDeploymentRef("%s-gateway-000", testName).
		StatusServiceRef("%s-gateway-000", testName).
		StatusAddress("http://%s-gateway-000.%s.svc.cluster.local", testName, testNamespace)

	serviceCreate := factories.Service().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.GenerateName("%s-gateway-", testName)
			om.AddLabel(streamingv1alpha1.GatewayLabelKey, testName)
			om.ControlledBy(gateway, scheme)
		}).
		AddSelectorLabel(streamingv1alpha1.GatewayLabelKey, testName).
		Ports(ports...)
	serviceGiven := serviceCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Name("%s-gateway-000", testName)
			om.Created(1)
		}).
		ClusterIP("10.10.10.10")

	deploymentCreate := factories.Deployment().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.GenerateName("%s-gateway-", testName)
			om.AddLabel(streamingv1alpha1.GatewayLabelKey, testName)
			om.ControlledBy(gateway, scheme)
		}).
		AddSelectorLabel(streamingv1alpha1.GatewayLabelKey, testName).
		PodTemplateSpec(func(pts factories.PodTemplateSpec) {
			pts.ContainerNamed("test", func(c *corev1.Container) {
				c.Image = "scratch"
			})
		})
	deploymentGiven := deploymentCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Name("%s-gateway-000", testName)
			om.Created(1)
		})

	table := rtesting.Table{{
		Name: "gateway does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted gateway",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			gateway.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
	}, {
		Name: "error fetching gateway",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Gateway"),
		},
		GivenObjects: []rtesting.Factory{
			gateway,
		},
		ShouldErr: true,
	}, {
		Name: "squatting gateway",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			gatewayMinimal,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			gatewayMinimal.
				StatusObservedGeneration(1).
				StatusConditions(
					gatewayConditionDeploymentReady.Unknown(),
					gatewayConditionReady.Unknown(),
					gatewayConditionServiceReady.Unknown(),
				),
		},
	}, {
		Name: "create service",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			gatewayMinimal.
				Ports(ports...),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "Created",
				`Created Service "%s-gateway-001"`, testName),
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			serviceCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			gatewayMinimal.
				StatusObservedGeneration(1).
				StatusConditions(
					gatewayConditionDeploymentReady.Unknown(),
					gatewayConditionReady.Unknown(),
					gatewayConditionServiceReady.True(),
				).
				StatusAddress("http://%s-gateway-001.%s.svc.cluster.local", testName, testNamespace).
				StatusServiceRef("%s-gateway-001", testName),
		},
	}, {
		Name: "create deployment",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			gatewayMinimal.
				Ports(ports...).
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.ContainerNamed("test", func(c *corev1.Container) {
						c.Image = "scratch"
					})
				}),
			serviceGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "Created",
				`Created Deployment "%s-gateway-001"`, testName),
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			gatewayMinimal.
				StatusObservedGeneration(1).
				StatusConditions(
					gatewayConditionDeploymentReady.Unknown(),
					gatewayConditionReady.Unknown(),
					gatewayConditionServiceReady.True(),
				).
				StatusAddress("http://%s-gateway-000.%s.svc.cluster.local", testName, testNamespace).
				StatusServiceRef("%s-gateway-000", testName).
				StatusDeploymentRef("%s-gateway-001", testName),
		},
	}, {
		Name: "update service and deployment",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			gatewayMinimal.
				Ports(ports...).
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.ContainerNamed("test", func(c *corev1.Container) {
						c.Image = "scratch"
					})
				}),
			serviceGiven.
				Ports(),
			deploymentGiven.
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.ContainerNamed("test", func(c *corev1.Container) {
						c.Image = "blah"
					})
				}),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Service "%s-gateway-000"`, testName),
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Deployment "%s-gateway-000"`, testName),
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			serviceGiven,
			deploymentGiven,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			gatewayMinimal.
				StatusObservedGeneration(1).
				StatusConditions(
					gatewayConditionDeploymentReady.Unknown(),
					gatewayConditionReady.Unknown(),
					gatewayConditionServiceReady.True(),
				).
				StatusAddress("http://%s-gateway-000.%s.svc.cluster.local", testName, testNamespace).
				StatusServiceRef("%s-gateway-000", testName).
				StatusDeploymentRef("%s-gateway-000", testName),
		},
	}, {
		Name: "ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			gatewayMinimal.
				Ports(ports...).
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.ContainerNamed("test", func(c *corev1.Container) {
						c.Image = "scratch"
					})
				}),
			serviceGiven,
			deploymentGiven.
				StatusConditions(
					deploymentConditionAvailable.True(),
					deploymentConditionProgressing.True(),
				),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			gatewayMinimal.
				StatusObservedGeneration(1).
				StatusConditions(
					gatewayConditionDeploymentReady.True(),
					gatewayConditionReady.True(),
					gatewayConditionServiceReady.True(),
				).
				StatusAddress("http://%s-gateway-000.%s.svc.cluster.local", testName, testNamespace).
				StatusServiceRef("%s-gateway-000", testName).
				StatusDeploymentRef("%s-gateway-000", testName),
		},
	}, {
		Name: "not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			gatewayMinimal.
				Ports(ports...).
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.ContainerNamed("test", func(c *corev1.Container) {
						c.Image = "scratch"
					})
				}),
			serviceGiven,
			deploymentGiven.
				StatusConditions(
					deploymentConditionAvailable.False().Reason("TestReason", "a human readable message"),
					deploymentConditionProgressing.Unknown(),
				),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(gateway, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			gatewayMinimal.
				StatusObservedGeneration(1).
				StatusConditions(
					gatewayConditionDeploymentReady.False().Reason("TestReason", "a human readable message"),
					gatewayConditionReady.False().Reason("TestReason", "a human readable message"),
					gatewayConditionServiceReady.True(),
				).
				StatusAddress("http://%s-gateway-000.%s.svc.cluster.local", testName, testNamespace).
				StatusServiceRef("%s-gateway-000", testName).
				StatusDeploymentRef("%s-gateway-000", testName),
		},
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		return streaming.GatewayReconciler(
			controllers.Config{
				Client:    client,
				APIReader: apiReader,
				Recorder:  recorder,
				Log:       log,
				Scheme:    scheme,
				Tracker:   tracker,
			},
		)
	})
}
