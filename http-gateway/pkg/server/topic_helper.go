package server

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	riffcs "github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/rest"
)

type TopicHelper interface {
	TopicExists(topicName string) (bool, error)
}

type topicHelper struct {
	client *riffcs.Clientset
}

func NewTopicHelper() (*topicHelper, error) {
	restConf, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	riffClient, err := riffcs.NewForConfig(restConf)
	if err != nil {
		return nil, err
	}

	return &topicHelper{client: riffClient}, nil
}

func (tw *topicHelper) TopicExists(topicName string) (bool, error) {
	_, err := tw.client.ProjectriffV1alpha1().Topics("default").Get(topicName, v1.GetOptions{})

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
