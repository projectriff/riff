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
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
)

// CredentialReconciler reconciles a Credential object
type CredentialReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
}

// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=build.projectriff.io,resources=applications;functions,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

func (r *CredentialReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("serviceaccount", req.NamespacedName)

	if req.Name != riffBuildServiceAccount {
		// ignore other service accounts, should never get here
		return ctrl.Result{}, nil
	}

	var originalSerivceAccount corev1.ServiceAccount
	if err := r.Get(ctx, req.NamespacedName, &originalSerivceAccount); err != nil && !apierrs.IsNotFound(err) {
		log.Error(err, "unable to fetch ServiceAccount")
		return ctrl.Result{}, err
	}

	// Don't modify the informers copy
	var serviceAccount corev1.ServiceAccount
	if &originalSerivceAccount != nil {
		serviceAccount = *(originalSerivceAccount.DeepCopy())
	}

	return r.reconcile(ctx, log, &serviceAccount, req.Namespace)
}

func (r *CredentialReconciler) reconcile(ctx context.Context, log logr.Logger, serviceAccount *corev1.ServiceAccount, namespace string) (ctrl.Result, error) {
	if serviceAccount != nil && serviceAccount.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	secretNames := sets.NewString()
	var secrets corev1.SecretList
	if err := r.List(ctx, &secrets, client.InNamespace(namespace), MatchingLabels(buildv1alpha1.CredentialLabelKey)); err != nil {
		log.Error(err, "Failed to get Secrets", "serviceaccount", serviceAccount)
		return ctrl.Result{Requeue: true}, err
	}
	for _, secret := range secrets.Items {
		secretNames.Insert(secret.Name)
	}

	if serviceAccount.Name == "" {
		if needed, err := r.isServiceAccountNeeded(ctx, secretNames, namespace); err != nil {
			return ctrl.Result{}, err
		} else if needed {
			serviceAccount, err := r.createServiceAccount(ctx, log, secretNames, namespace)
			if err != nil {
				log.Error(err, "Failed to create ServiceAccount", "serviceaccount", serviceAccount)
				return ctrl.Result{}, err
			}
		}
	} else {
		serviceAccount, err := r.reconcileServiceAccount(ctx, log, serviceAccount, secretNames)
		if err != nil {
			log.Error(err, "Failed to reconcile ServiceAccount", "serviceaccount", serviceAccount)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *CredentialReconciler) reconcileServiceAccount(ctx context.Context, log logr.Logger, existingServiceAccount *corev1.ServiceAccount, desiredBoundSecrets sets.String) (*corev1.ServiceAccount, error) {
	serviceAccount := existingServiceAccount.DeepCopy()
	boundSecrets := sets.NewString(strings.Split(serviceAccount.Annotations[buildv1alpha1.CredentialsAnnotationKey], ",")...)
	removeSecrets := boundSecrets.Difference(desiredBoundSecrets)

	secrets := []corev1.ObjectReference{}
	// filter out secrets no longer bound
	for _, secret := range serviceAccount.Secrets {
		if !removeSecrets.Has(secret.Name) {
			secrets = append(secrets, secret)
		}
	}
	// add new secrets
	for _, secret := range desiredBoundSecrets.Difference(boundSecrets).List() {
		secrets = append(secrets, corev1.ObjectReference{Name: secret})
	}
	serviceAccount.Secrets = secrets

	if serviceAccount.Annotations == nil {
		serviceAccount.Annotations = map[string]string{}
	}
	serviceAccount.Annotations[buildv1alpha1.CredentialsAnnotationKey] = strings.Join(desiredBoundSecrets.List(), ",")

	if serviceAccountSemanticEquals(serviceAccount, existingServiceAccount) {
		// No differences to reconcile.
		return serviceAccount, nil
	}

	log.Info("reconciling serviceaccount", "diff", cmp.Diff(existingServiceAccount.Secrets, serviceAccount.Secrets))
	return serviceAccount, r.Update(ctx, serviceAccount)
}

func (r *CredentialReconciler) isServiceAccountNeeded(ctx context.Context, secretNames sets.String, namespace string) (bool, error) {
	if secretNames.Len() != 0 {
		return true, nil
	}
	var applications buildv1alpha1.ApplicationList
	if err := r.List(ctx, &applications, client.InNamespace(namespace)); err != nil {
		return false, err
	} else if len(applications.Items) != 0 {
		return true, nil
	}
	var functions buildv1alpha1.FunctionList
	if err := r.List(ctx, &functions, client.InNamespace(namespace)); err != nil {
		return false, err
	} else if len(functions.Items) != 0 {
		return true, nil
	}
	var containers buildv1alpha1.ContainerList
	if err := r.List(ctx, &containers, client.InNamespace(namespace)); err != nil {
		return false, err
	} else if len(containers.Items) != 0 {
		return true, nil
	}
	return false, nil
}

func (r *CredentialReconciler) createServiceAccount(ctx context.Context, log logr.Logger, secretNames sets.String, namespace string) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      riffBuildServiceAccount,
			Namespace: namespace,
			Annotations: map[string]string{
				buildv1alpha1.CredentialsAnnotationKey: strings.Join(secretNames.List(), ","),
			},
		},
		Secrets: make([]corev1.ObjectReference, secretNames.Len()),
	}
	for i, secretName := range secretNames.UnsortedList() {
		serviceAccount.Secrets[i] = corev1.ObjectReference{Name: secretName}
	}
	log.Info("creating serviceaccount", "secrets", serviceAccount.Secrets)
	return serviceAccount, r.Create(ctx, serviceAccount)
}

func serviceAccountSemanticEquals(desiredServiceAccount, serviceAccount *corev1.ServiceAccount) bool {
	return equality.Semantic.DeepEqual(desiredServiceAccount.Secrets, serviceAccount.Secrets) &&
		equality.Semantic.DeepEqual(desiredServiceAccount.Annotations, serviceAccount.Annotations)
}

// MatchingLabels filters the list/delete operation for a given LabelSelctor
type MatchingLabels string

func (m MatchingLabels) ApplyToList(opts *client.ListOptions) {
	sel, _ := labels.Parse(string(m))
	opts.LabelSelector = sel
}

func (r *CredentialReconciler) SetupWithManager(mgr ctrl.Manager) error {
	enqueueServiceAccountForCredential := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			if _, ok := a.Meta.GetLabels()[buildv1alpha1.CredentialLabelKey]; !ok {
				// not all secrets are credentials
				return []reconcile.Request{}
			}
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Namespace: a.Meta.GetNamespace(),
						Name:      riffBuildServiceAccount,
					},
				},
			}
		}),
	}

	enqueueServiceAccountForBuild := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Namespace: a.Meta.GetNamespace(),
						Name:      riffBuildServiceAccount,
					},
				},
			}
		}),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				sa, ok := e.Object.(*corev1.ServiceAccount)
				if !ok {
					// not a serviceaccount, allow
					return true
				}
				// filter services accounts to only be the riff-build sa
				return sa.Name == riffBuildServiceAccount
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				sa, ok := e.ObjectNew.(*corev1.ServiceAccount)
				if !ok {
					// not a serviceaccount, allow
					return true
				}
				// filter services accounts to only be the riff-build sa
				return sa.Name == riffBuildServiceAccount
			},
		}).
		// watch for secret mutations to bind to service account
		Watches(&source.Kind{Type: &corev1.Secret{}}, enqueueServiceAccountForCredential).
		// watch for build mutations to create service account
		Watches(&source.Kind{Type: &buildv1alpha1.Application{}}, enqueueServiceAccountForBuild).
		Watches(&source.Kind{Type: &buildv1alpha1.Function{}}, enqueueServiceAccountForBuild).
		Watches(&source.Kind{Type: &buildv1alpha1.Container{}}, enqueueServiceAccountForBuild).
		Complete(r)
}
