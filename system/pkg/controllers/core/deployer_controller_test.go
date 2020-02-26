/*
Copyright 2019 the original author or authors.

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

package core_test

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/projectriff/riff/system/pkg/apis/core/v1alpha1"
	"github.com/projectriff/riff/system/pkg/controllers"

	corecontrollers "github.com/projectriff/riff/system/pkg/controllers/core"
	rtesting "github.com/projectriff/riff/system/pkg/controllers/testing"
	"github.com/projectriff/riff/system/pkg/controllers/testing/factories"
	"github.com/projectriff/riff/system/pkg/tracker"
)

func TestDeployerReconciler(t *testing.T) {
	testNamespace := "test-namespace"
	testName := "test-deployer"
	testKey := types.NamespacedName{Namespace: testNamespace, Name: testName}
	testImagePrefix := "example.com/repo"
	testSha256 := "cf8b4c69d5460f88530e1c80b8856a70801f31c50b191c8413043ba9b160a43e"
	testImage := fmt.Sprintf("%s@sha256:%s", testImagePrefix, testSha256)
	testConditionReason := "TestReason"
	testConditionMessage := "meaningful, yet concise"
	testDomain := "example.com"
	testHost := fmt.Sprintf("%s.%s.%s", testName, testNamespace, testDomain)
	testURL := fmt.Sprintf("http://%s", testHost)
	testAddressURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", testName, testNamespace)
	testLabelKey := "test-label-key"
	testLabelValue := "test-label-value"

	deployerConditionDeploymentReady := factories.Condition().Type(corev1alpha1.DeployerConditionDeploymentReady)
	deployerConditionIngressReady := factories.Condition().Type(corev1alpha1.DeployerConditionIngressReady)
	deployerConditionReady := factories.Condition().Type(corev1alpha1.DeployerConditionReady)
	deployerConditionServiceReady := factories.Condition().Type(corev1alpha1.DeployerConditionServiceReady)

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = buildv1alpha1.AddToScheme(scheme)
	_ = corev1alpha1.AddToScheme(scheme)

	deployerMinimal := factories.DeployerCore().
		NamespaceName(testNamespace, testName)
	deployerValid := deployerMinimal.
		Image(testImage).
		IngressPolicy(corev1alpha1.IngressPolicyClusterLocal)

	deploymentCreate := factories.Deployment().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.GenerateName("%s-deployer-", deployerMinimal.Create().GetName())
			om.AddLabel(corev1alpha1.DeployerLabelKey, deployerMinimal.Create().GetName())
			om.ControlledBy(deployerMinimal, scheme)
		}).
		AddSelectorLabel(corev1alpha1.DeployerLabelKey, deployerMinimal.Create().GetName()).
		HandlerContainer(func(container *corev1.Container) {
			container.Image = testImage
			container.Ports = []corev1.ContainerPort{
				{Name: "http", ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
			}
			container.Env = []corev1.EnvVar{
				{Name: "PORT", Value: "8080"},
			}
			container.ReadinessProbe = &corev1.Probe{
				Handler: corev1.Handler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.FromInt(8080),
					},
				},
			}
		})
	deploymentGiven := deploymentCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Name("%s%s", om.Create().GenerateName, "000")
			om.Created(1)
		})

	serviceCreate := factories.Service().
		NamespaceName(testNamespace, testName).
		ObjectMeta(func(om factories.ObjectMeta) {
			om.AddLabel(corev1alpha1.DeployerLabelKey, deployerMinimal.Create().GetName())
			om.ControlledBy(deployerMinimal, scheme)
		}).
		AddSelectorLabel(corev1alpha1.DeployerLabelKey, deployerMinimal.Create().GetName()).
		Ports(
			corev1.ServicePort{
				Name:       "http",
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
		)
	serviceGiven := serviceCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Created(1)
			om.ControlledBy(deployerMinimal, scheme)
		})

	ingressCreate := factories.Ingress().
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Namespace(testNamespace)
			om.GenerateName("%s-deployer-", deployerMinimal.Create().GetName())
			om.AddLabel(corev1alpha1.DeployerLabelKey, deployerMinimal.Create().GetName())
			om.ControlledBy(deployerMinimal, scheme)
		}).
		HostToService(testHost, serviceGiven.Create().GetName())
	ingressGiven := ingressCreate.
		ObjectMeta(func(om factories.ObjectMeta) {
			om.Name("%s%s", om.Create().GenerateName, "000")
			om.Created(1)
		})

	testApplication := factories.Application().
		NamespaceName(testNamespace, "my-application").
		StatusLatestImage(testImage)
	testFunction := factories.Function().
		NamespaceName(testNamespace, "my-function").
		StatusLatestImage(testImage)
	testContainer := factories.Container().
		NamespaceName(testNamespace, "my-container").
		StatusLatestImage(testImage)

	testSettings := factories.ConfigMap().
		NamespaceName("riff-system", "riff-core-settings").
		AddData("defaultDomain", "example.com")

	table := rtesting.Table{{
		Name: "deployer does not exist",
		Key:  testKey,
	}, {
		Name: "ignore deleted deployer",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerValid.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.Deleted(1)
				}),
		},
	}, {
		Name: "deployer get error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("get", "Deployer"),
		},
		ShouldErr: true,
	}, {
		Name: "create resources, from application",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				ApplicationRef(testApplication.Create().GetName()),
			testApplication,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Deployment "%s-deployer-001"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Service "%s"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
			serviceCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-001", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "create resources, from application, application not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				ApplicationRef(testApplication.Create().GetName()),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, deployerMinimal, scheme),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.Unknown(),
				),
		},
	}, {
		Name: "create resources, from application, no latest image",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				ApplicationRef(testApplication.Create().GetName()),
			testApplication.
				StatusLatestImage(""),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testApplication, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.Unknown(),
				),
		},
	}, {
		Name: "create resources, from function",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				FunctionRef(testFunction.Create().GetName()),
			testFunction,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Deployment "%s-deployer-001"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Service "%s"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
			serviceCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-001", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "create resources, from function, function not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				FunctionRef(testFunction.Create().GetName()),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, deployerMinimal, scheme),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.Unknown(),
				),
		},
	}, {
		Name: "create resources, from function, no latest image",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				FunctionRef(testFunction.Create().GetName()),
			testFunction.
				StatusLatestImage(""),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testFunction, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.Unknown(),
				),
		},
	}, {
		Name: "create resources, from container",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				ContainerRef(testContainer.Create().GetName()),
			testContainer,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Deployment "%s-deployer-001"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Service "%s"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
			serviceCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-001", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "create resources, from container, container not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				ContainerRef(testContainer.Create().GetName()),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.Unknown(),
				),
		},
	}, {
		Name: "create resources, from container, no latest image",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				ContainerRef(testContainer.Create().GetName()),
			testContainer.
				StatusLatestImage(""),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testContainer, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.Unknown(),
				),
		},
	}, {
		Name: "create resources, from image",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Deployment "%s-deployer-001"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Service "%s"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
			serviceCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-001", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "create deployment, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Deployment"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Deployment "": inducing failure for create Deployment`),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.Unknown(),
				).
				StatusLatestImage(testImage),
		},
	}, {
		Name: "create service, error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Service"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Deployment "%s-deployer-001"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Service "%s": inducing failure for create Service`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
			serviceCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.Unknown(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-001", deployerMinimal.Create().GetName()),
		},
	}, {
		Name: "create service, conflicted",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Service", rtesting.InduceFailureOpts{
				Error: apierrs.NewAlreadyExists(schema.GroupResource{}, testName),
			}),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Deployment "%s-deployer-001"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Service "%s":  "%s" already exists`, testName, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
			serviceCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.False().Reason("NotOwned", `There is an existing Service "test-deployer" that the Deployer does not own.`),
					deployerConditionServiceReady.False().Reason("NotOwned", `There is an existing Service "test-deployer" that the Deployer does not own.`),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-001", deployerMinimal.Create().GetName()),
		},
	}, {
		Name: "update deployment",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven.
				HandlerContainer(func(container *corev1.Container) {
					// change to reverse
					container.Env = nil
				}),
			serviceGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Deployment "%s"`, deploymentGiven.Create().GetName()),
		},
		ExpectUpdates: []rtesting.Factory{
			deploymentGiven,
		},
	}, {
		Name: "update deployment, update error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Deployment"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven.
				HandlerContainer(func(container *corev1.Container) {
					// change to reverse
					container.Env = nil
				}),
			serviceGiven,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "UpdateFailed",
				`Failed to update Deployment "%s": inducing failure for update Deployment`, deploymentGiven.Create().GetName()),
		},
		ExpectUpdates: []rtesting.Factory{
			deploymentGiven,
		},
	}, {
		Name: "update deployment, list deployments failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "DeploymentList"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven.
				HandlerContainer(func(container *corev1.Container) {
					// change to reverse
					container.Env = nil
				}),
			serviceGiven,
		},
		ShouldErr: true,
	}, {
		Name: "update service",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven,
			serviceGiven.
				// change to reverse
				Ports(),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Service "%s"`, serviceGiven.Create().GetName()),
		},
		ExpectUpdates: []rtesting.Factory{
			serviceGiven,
		},
	}, {
		Name: "update service, update error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Service"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven,
			serviceGiven.
				// change to reverse
				Ports(),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "UpdateFailed",
				`Failed to update Service "%s": inducing failure for update Service`, serviceGiven.Create().GetName()),
		},
		ExpectUpdates: []rtesting.Factory{
			serviceGiven,
		},
	}, {
		Name: "update service, list services failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "ServiceList"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven,
			serviceGiven.
				// change to reverse
				Ports(),
		},
		ShouldErr: true,
	}, {
		Name: "cleanup extra deployments",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-001", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven.
				NamespaceName(testNamespace, "extra-deployment-1"),
			deploymentGiven.
				NamespaceName(testNamespace, "extra-deployment-2"),
			serviceGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Deployment "%s"`, "extra-deployment-1"),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Deployment "%s"`, "extra-deployment-2"),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Deployment "%s-deployer-001"`, testName),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "apps", Kind: "Deployment", Namespace: testNamespace, Name: "extra-deployment-1"},
			{Group: "apps", Kind: "Deployment", Namespace: testNamespace, Name: "extra-deployment-2"},
		},
		ExpectCreates: []rtesting.Factory{
			deploymentCreate,
		},
	}, {
		Name: "cleanup extra deployments, delete deployment failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("delete", "Deployment"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-001", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven.
				NamespaceName(testNamespace, "extra-deployment-1"),
			deploymentGiven.
				NamespaceName(testNamespace, "extra-deployment-2"),
			serviceGiven,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "DeleteFailed",
				`Failed to delete Deployment "%s": inducing failure for delete Deployment`, "extra-deployment-1"),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "apps", Kind: "Deployment", Namespace: testNamespace, Name: "extra-deployment-1"},
		},
	}, {
		Name: "cleanup extra services",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven,
			serviceGiven.
				NamespaceName(testNamespace, "extra-service-1"),
			serviceGiven.
				NamespaceName(testNamespace, "extra-service-2"),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Service "%s"`, "extra-service-1"),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Service "%s"`, "extra-service-2"),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Service "%s"`, testName),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Kind: "Service", Namespace: testNamespace, Name: "extra-service-1"},
			{Kind: "Service", Namespace: testNamespace, Name: "extra-service-2"},
		},
		ExpectCreates: []rtesting.Factory{
			serviceCreate,
		},
	}, {
		Name: "cleanup extra services, delete service failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("delete", "Service"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
			deploymentGiven,
			serviceGiven.
				NamespaceName(testNamespace, "extra-service-1"),
			serviceGiven.
				NamespaceName(testNamespace, "extra-service-2"),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "DeleteFailed",
				`Failed to delete Service "%s": inducing failure for delete Service`, "extra-service-1"),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Kind: "Service", Namespace: testNamespace, Name: "extra-service-1"},
		},
	}, {
		Name: "create ingress",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven,
			serviceGiven,
			testSettings,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testSettings, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Ingress "%s-deployer-001"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			ingressCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusIngressRef("%s-deployer-001", testName).
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
	}, {
		Name: "create ingress, create failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("create", "Ingress"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven,
			serviceGiven,
			testSettings,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testSettings, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "CreationFailed",
				`Failed to create Ingress "": inducing failure for create Ingress`),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectCreates: []rtesting.Factory{
			ingressCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL(testAddressURL),
		},
	}, {
		Name: "delete ingress",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyClusterLocal),
			deploymentGiven,
			serviceGiven,
			ingressGiven,
			testSettings.
				AddData("defaultDomain", "not.example.com"),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Ingress "%s"`, ingressGiven.Create().GetName()),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "networking.k8s.io", Kind: "Ingress", Namespace: testNamespace, Name: ingressGiven.Create().GetName()},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL(testAddressURL),
		},
	}, {
		Name: "delete ingress, delete failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("delete", "Ingress"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyClusterLocal),
			deploymentGiven,
			serviceGiven,
			ingressGiven,
			testSettings.
				AddData("defaultDomain", "not.example.com"),
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "DeleteFailed",
				`Failed to delete Ingress "%s": inducing failure for delete Ingress`, ingressGiven.Create().GetName()),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "networking.k8s.io", Kind: "Ingress", Namespace: testNamespace, Name: ingressGiven.Create().GetName()},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL(testAddressURL),
		},
	}, {
		Name: "update ingress",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven,
			serviceGiven,
			ingressGiven,
			testSettings.
				AddData("defaultDomain", "not.example.com"),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testSettings, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Ingress "%s"`, ingressGiven.Create().GetName()),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			ingressGiven.
				HostToService(fmt.Sprintf("%s.%s.%s", testName, testNamespace, "not.example.com"), serviceGiven.Create().GetName()),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusIngressRef("%s-deployer-000", testName).
				StatusAddressURL(testAddressURL).
				StatusURL("http://%s.%s.%s", testName, testNamespace, "not.example.com"),
		},
	}, {
		Name: "update ingress, update failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Ingress"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven,
			serviceGiven,
			ingressGiven,
			testSettings.
				AddData("defaultDomain", "not.example.com"),
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testSettings, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "UpdateFailed",
				`Failed to update Ingress "%s": inducing failure for update Ingress`, ingressGiven.Create().GetName()),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectUpdates: []rtesting.Factory{
			ingressGiven.
				HostToService(fmt.Sprintf("%s.%s.%s", testName, testNamespace, "not.example.com"), serviceGiven.Create().GetName()),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL(testAddressURL),
		},
	}, {
		Name: "remove extra ingress",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven,
			serviceGiven,
			ingressGiven.
				NamespaceName(testNamespace, "extra-ingress-1"),
			ingressGiven.
				NamespaceName(testNamespace, "extra-ingress-2"),
			testSettings,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Ingress "%s"`, "extra-ingress-1"),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Deleted",
				`Deleted Ingress "%s"`, "extra-ingress-2"),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Created",
				`Created Ingress "%s-deployer-001"`, testName),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testSettings, deployerMinimal, scheme),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "networking.k8s.io", Kind: "Ingress", Namespace: testNamespace, Name: "extra-ingress-1"},
			{Group: "networking.k8s.io", Kind: "Ingress", Namespace: testNamespace, Name: "extra-ingress-2"},
		},
		ExpectCreates: []rtesting.Factory{
			ingressCreate,
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusIngressRef("%s-deployer-001", testName).
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
	}, {
		Name: "remove extra ingress, listing failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("list", "IngressList"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven,
			serviceGiven,
			ingressGiven.
				NamespaceName(testNamespace, "extra-ingress-1"),
			ingressGiven.
				NamespaceName(testNamespace, "extra-ingress-2"),
			testSettings,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "remove extra ingress, delete failed",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("delete", "Ingress"),
		},
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				Image(testImage).
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven,
			serviceGiven,
			ingressGiven.
				NamespaceName(testNamespace, "extra-ingress-1"),
			ingressGiven.
				NamespaceName(testNamespace, "extra-ingress-2"),
			testSettings,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "DeleteFailed",
				`Failed to delete Ingress "%s": inducing failure for delete Ingress`, "extra-ingress-1"),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectDeletes: []rtesting.DeleteRef{
			{Group: "networking.k8s.io", Kind: "Ingress", Namespace: testNamespace, Name: "extra-ingress-1"},
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "propagate labels",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerMinimal.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel(testLabelKey, testLabelValue)
				}).
				IngressPolicy(corev1alpha1.IngressPolicyExternal).
				Image(testImage).
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusIngressRef(ingressGiven.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()).
				StatusURL(testURL),
			deploymentGiven,
			serviceGiven,
			ingressGiven,
			testSettings,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testSettings, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Deployment "%s"`, deploymentGiven.Create().GetName()),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Service "%s"`, serviceGiven.Create().GetName()),
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "Updated",
				`Updated Ingress "%s"`, ingressGiven.Create().GetName()),
		},
		ExpectUpdates: []rtesting.Factory{
			deploymentGiven.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel(testLabelKey, testLabelValue)
				}).
				PodTemplateSpec(func(pts factories.PodTemplateSpec) {
					pts.AddLabel(testLabelKey, testLabelValue)
				}),
			serviceGiven.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel(testLabelKey, testLabelValue)
				}),
			ingressGiven.
				ObjectMeta(func(om factories.ObjectMeta) {
					om.AddLabel(testLabelKey, testLabelValue)
				}),
		},
	}, {
		Name: "ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerValid,
			deploymentGiven.
				StatusConditions(
					factories.Condition().Type("Available").True(),
					factories.Condition().Type("Progressing").Unknown(),
				),
			serviceGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.True(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.True(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "ready, with ingress",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerValid.
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven.
				StatusConditions(
					factories.Condition().Type("Available").True(),
					factories.Condition().Type("Progressing").Unknown(),
				),
			serviceGiven,
			ingressGiven.
				StatusLoadBalancer(
					corev1.LoadBalancerIngress{
						Hostname: testHost,
					},
				),
			testSettings,
		},
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testSettings, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.True(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.True(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusIngressRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusAddressURL(testAddressURL).
				StatusURL(testURL),
		},
	}, {
		Name: "not ready",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerValid,
			deploymentGiven.
				StatusConditions(
					factories.Condition().Type("Available").False().Reason(testConditionReason, testConditionMessage),
					factories.Condition().Type("Progressing").Unknown(),
				),
			serviceGiven,
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.False().Reason(testConditionReason, testConditionMessage),
					deployerConditionIngressReady.True(),
					deployerConditionReady.False().Reason(testConditionReason, testConditionMessage),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "update status error",
		Key:  testKey,
		WithReactors: []rtesting.ReactionFunc{
			rtesting.InduceFailure("update", "Deployer"),
		},
		GivenObjects: []rtesting.Factory{
			deployerValid,
			deploymentGiven,
			serviceGiven,
		},
		ShouldErr: true,
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeWarning, "StatusUpdateFailed",
				`Failed to update status: inducing failure for update Deployer`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.True(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}, {
		Name: "settings not found",
		Key:  testKey,
		GivenObjects: []rtesting.Factory{
			deployerValid.
				IngressPolicy(corev1alpha1.IngressPolicyExternal),
			deploymentGiven,
			serviceGiven,
			ingressGiven,
		},
		ShouldErr: true,
		ExpectTracks: []rtesting.TrackRequest{
			rtesting.NewTrackRequest(testSettings, deployerMinimal, scheme),
		},
		ExpectEvents: []rtesting.Event{
			rtesting.NewEvent(deployerMinimal, scheme, corev1.EventTypeNormal, "StatusUpdated",
				`Updated status`),
		},
		ExpectStatusUpdates: []rtesting.Factory{
			deployerMinimal.
				StatusConditions(
					deployerConditionDeploymentReady.Unknown(),
					deployerConditionIngressReady.Unknown(),
					deployerConditionReady.Unknown(),
					deployerConditionServiceReady.True(),
				).
				StatusLatestImage(testImage).
				StatusDeploymentRef("%s-deployer-000", deployerMinimal.Create().GetName()).
				StatusServiceRef(deployerMinimal.Create().GetName()).
				StatusAddressURL("http://%s.%s.svc.cluster.local", serviceCreate.Create().GetName(), serviceCreate.Create().GetNamespace()),
		},
	}}

	table.Test(t, scheme, func(t *testing.T, row *rtesting.Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler {
		return corecontrollers.DeployerReconciler(
			controllers.Config{
				Client:    client,
				APIReader: apiReader,
				Recorder:  recorder,
				Scheme:    scheme,
				Log:       log,
				Tracker:   tracker,
			},
		)
	})
}
