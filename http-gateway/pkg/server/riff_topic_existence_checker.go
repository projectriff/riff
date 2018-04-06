package server

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	riffcs "github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/client-go/rest"
)

const (
	defaultNamespace = "default" // expected to be used by consumers of TopicExists
)

// RiffTopicExistenceChecker allows the http-gateway to check for the existence of
// a Riff Topic before attempting to send a message to that topic.
type RiffTopicExistenceChecker interface {
	TopicExists(namespace string, topicName string) (bool, error)
}

type riffTopicExistenceChecker struct {
	client *riffcs.Clientset
}

func NewRiffTopicExistenceChecker() (*riffTopicExistenceChecker, error) {
	restConf, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	riffClient, err := riffcs.NewForConfig(restConf)
	if err != nil {
		return nil, err
	}

	return &riffTopicExistenceChecker{client: riffClient}, nil
}

// TopicExists checks to see if Kubernetes is aware of a Riff Topic in a namespace.
// If the topic exists, it returns (true, nil)
// If the topic does not exist, it returns (false, nil)
// If there is an unexpected error, it returns (false, err), where 'err' is the unexpected
// error that was encountered.
func (tec *riffTopicExistenceChecker) TopicExists(namespace string, topicName string) (bool, error) {
	_, err := tec.client.ProjectriffV1alpha1().Topics(namespace).Get(topicName, v1.GetOptions{})

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
