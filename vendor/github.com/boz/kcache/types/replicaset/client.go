package replicaset

import (
	"github.com/boz/kcache/client"
	"k8s.io/client-go/kubernetes"
)

const resourceName = "replicasets"

func NewClient(cs kubernetes.Interface, ns string) client.Client {
	scope := cs.ExtensionsV1beta1()
	return client.ForResource(scope.RESTClient(), resourceName, ns)
}
