package join

import (
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ metav1.Object
var _ corev1.Pod
var _ corev1.Secret
var _ corev1.Service
var _ corev1.Event
var _ corev1.Node
var _ corev1.ReplicationController
var _ extv1beta1.Deployment
var _ extv1beta1.Ingress
var _ extv1beta1.ReplicaSet
var _ extv1beta1.DaemonSet
