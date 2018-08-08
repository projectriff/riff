package node

import (
	"github.com/boz/kcache/client"
	"k8s.io/client-go/kubernetes"
)

const resourceName = "nodes"

func NewClient(cs kubernetes.Interface, ns string) client.Client {
	scope := cs.CoreV1()
	return client.ForResource(scope.RESTClient(), resourceName, ns)
}
