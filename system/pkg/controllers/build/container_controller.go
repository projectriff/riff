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

package build

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	gauthn "github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	"github.com/projectriff/system/pkg/authn"
)

// ContainerReconciler reconciles a Container object
type ContainerReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

var containerPollingInterval = 1 * time.Minute

// +kubebuilder:rbac:groups=build.projectriff.io,resources=containers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.projectriff.io,resources=containers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func (r *ContainerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("container", req.NamespacedName)

	var originalContainer buildv1alpha1.Container
	if err := r.Get(ctx, req.NamespacedName, &originalContainer); err != nil {
		if apierrs.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Container")
		return ctrl.Result{}, err
	}
	container := *(originalContainer.DeepCopy())

	container.Default()
	container.Status.InitializeConditions()

	result, err := r.reconcile(ctx, log, &container)

	// check if status has changed before updating, unless requeued
	if !equality.Semantic.DeepEqual(container.Status, originalContainer.Status) && container.GetDeletionTimestamp() == nil {
		// update status
		log.Info("updating container status", "diff", cmp.Diff(originalContainer.Status, container.Status))
		if updateErr := r.Status().Update(ctx, &container); updateErr != nil {
			log.Error(updateErr, "unable to update Container status", "container", container)
			r.Recorder.Eventf(&container, corev1.EventTypeWarning, "StatusUpdateFailed",
				"Failed to update status: %v", updateErr)
			return ctrl.Result{Requeue: true}, updateErr
		}
		r.Recorder.Eventf(&container, corev1.EventTypeNormal, "StatusUpdated",
			"Updated status")
	}

	// return original reconcile result
	return result, err
}

func (r *ContainerReconciler) reconcile(ctx context.Context, log logr.Logger, container *buildv1alpha1.Container) (ctrl.Result, error) {
	if container.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	// resolve target image
	targetImageRef, err := r.resolveTargetImage(ctx, log, container)
	if err != nil {
		if err == errMissingDefaultPrefix {
			container.Status.MarkImageDefaultPrefixMissing(err.Error())
		} else {
			container.Status.MarkImageInvalid(err.Error())
		}
		return ctrl.Result{}, err
	}
	container.Status.TargetImage = targetImageRef.Name()

	latestImage, err := r.resolveDigestReference(ctx, log, targetImageRef, container)
	if err != nil {
		container.Status.MarkImageInvalid(err.Error())
		return ctrl.Result{}, err
	}

	container.Status.MarkImageResolved()

	container.Status.LatestImage = latestImage

	container.Status.ObservedGeneration = container.Generation

	return ctrl.Result{
		RequeueAfter: containerPollingInterval,
	}, nil
}

func (r *ContainerReconciler) resolveTargetImage(ctx context.Context, log logr.Logger, container *buildv1alpha1.Container) (name.Reference, error) {
	image := container.Spec.Image
	var err error
	if strings.HasPrefix(container.Spec.Image, "_") {
		image, err = r.interpolatePrefix(ctx, log, container)
		if err != nil {
			return nil, err
		}
	}

	ref, err := name.ParseReference(image)
	if err != nil {
		log.Error(err, "invalid target image reference", "image", image)
		return nil, err
	}
	return ref, nil
}

func (r *ContainerReconciler) interpolatePrefix(ctx context.Context, log logr.Logger, container *buildv1alpha1.Container) (string, error) {
	var riffBuildConfig corev1.ConfigMap
	if err := r.Get(ctx, types.NamespacedName{Namespace: container.Namespace, Name: riffBuildServiceAccount}, &riffBuildConfig); err != nil {
		if apierrs.IsNotFound(err) {
			return "", errMissingDefaultPrefix
		}
		return "", err
	}
	defaultPrefix := riffBuildConfig.Data["default-image-prefix"]
	if defaultPrefix == "" {
		return "", errMissingDefaultPrefix
	}
	image, err := buildv1alpha1.ResolveDefaultImage(container, defaultPrefix)
	if err != nil {
		return "", err
	}
	return image, nil
}

func (r *ContainerReconciler) resolveDigestReference(ctx context.Context, log logr.Logger, ref name.Reference, container *buildv1alpha1.Container) (string, error) {
	keychain, err := r.constructKeychain(ctx, log, container)
	if err != nil {
		return "", err
	}

	auth, err := keychain.Resolve(ref.Context().Registry)
	if err != nil {
		log.Error(err, "unable to resolve auth for registry", "registry", ref.Context().RegistryStr())
		return "", err
	}

	img, err := remote.Image(ref, remote.WithAuth(auth))
	if err != nil {
		log.Error(err, "failed to read image", "image", ref.String())
		return "", err
	}

	digest, err := img.Digest()
	if err != nil {
		log.Error(err, "failed to get image digest", "image", ref.String())
		return "", err
	}

	return fmt.Sprintf("%s@%s", ref.Context().Name(), digest), nil
}

func (r *ContainerReconciler) constructKeychain(ctx context.Context, log logr.Logger, container *buildv1alpha1.Container) (gauthn.Keychain, error) {
	var serviceAccount corev1.ServiceAccount
	if err := r.Get(ctx, types.NamespacedName{Namespace: container.Namespace, Name: riffBuildServiceAccount}, &serviceAccount); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("service account not found", "service-account", riffBuildServiceAccount)
			return nil, err
		} else {
			log.Error(err, "failed to get service account", "service-account", riffBuildServiceAccount)
			return nil, err
		}
	}
	secrets, err := r.fetchSecrets(serviceAccount, ctx, log)
	if err != nil {
		return nil, err
	}

	return gauthn.NewMultiKeychain(authn.NewSecretsKeychain(secrets), gauthn.DefaultKeychain), nil
}

func (r *ContainerReconciler) fetchSecrets(serviceAccount corev1.ServiceAccount, ctx context.Context, log logr.Logger) ([]corev1.Secret, error) {
	var secrets []corev1.Secret
	for _, secretRef := range serviceAccount.Secrets {
		var secret corev1.Secret
		if err := r.Get(ctx, types.NamespacedName{Namespace: serviceAccount.Namespace, Name: secretRef.Name}, &secret); err != nil {
			if apierrs.IsNotFound(err) {
				log.Info("secret not found", "secret", secretRef.Name)
				continue
			} else {
				log.Error(err, "failed to get secret", "secret", secretRef.Name)
				return nil, err
			}
		}
		secrets = append(secrets, secret)
	}
	return secrets, nil
}

func (r *ContainerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&buildv1alpha1.Container{}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, handler.Funcs{}).
		Watches(&source.Kind{Type: &corev1.ServiceAccount{}}, handler.Funcs{}).
		Complete(r)
}
