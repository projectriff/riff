package devutil

import (
	"errors"
	"io/ioutil"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var (
	StreamGVR = schema.GroupVersionResource{
		Group:    "streaming.projectriff.io",
		Version:  "v1alpha1",
		Resource: "streams",
	}
	SecretGVR = schema.GroupVersionResource{
		Version:  "v1",
		Resource: "secrets",
	}
)

const namespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

type K8sClient struct {
	dc dynamic.Interface
}

func NewK8sClient() *K8sClient {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &K8sClient{
		dc: dynamicClient,
	}
}

func (c *K8sClient) GetNestedString(streamName, namespace string, gvr schema.GroupVersionResource, fields ...string) (string, error) {
	ns, err := resolveNamespace(namespace)
	if err != nil {
		return "", err
	}

	stream, err := c.dc.Resource(gvr).Namespace(ns).Get(streamName, v1.GetOptions{})
	if err != nil {
		return "", err
	}

	topic, found, err := unstructured.NestedString(stream.UnstructuredContent(), fields...)
	if err != nil {
		return "", err
	}
	if !found {
		return "", errors.New("unexpected structure of status")
	}
	return topic, nil
}

func resolveNamespace(namespace string) (string, error) {
	if namespace != "" {
		return namespace, nil
	}
	return getDefaultNamespace()
}
func getDefaultNamespace() (string, error) {
	namespacebytes, err := ioutil.ReadFile(namespaceFilePath)
	if err != nil {
		return "", err
	}
	return string(namespacebytes), nil
}
