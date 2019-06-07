package replicationcontroller

import (
	"github.com/boz/kcache/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const resourceName = string(corev1.ResourceReplicationControllers)

func NewClient(cs kubernetes.Interface, ns string) client.Client {
	scope := cs.CoreV1()
	return client.ForResource(scope.RESTClient(), resourceName, ns)
}
