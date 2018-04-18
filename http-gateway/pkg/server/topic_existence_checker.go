package server

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	defaultNamespace = "default" // expected to be used by consumers of TopicExists
)

// TopicExistenceChecker allows the http-gateway to check for the existence of
// a Topic before attempting to send a message to that topic.
type TopicExistenceChecker interface {
	TopicExists(namespace string, topicName string) (bool, error)
}

type riffTopicExistenceChecker struct {
	client *versioned.Clientset
}

type alwaysTrueTopicExistenceChecker struct {
}

// NewAlwaysTrueTopicExistenceChecker configures a TopicExistenceChecker that always returns true.
func NewAlwaysTrueTopicExistenceChecker() TopicExistenceChecker {
	return &alwaysTrueTopicExistenceChecker{}
}

// NewRiffTopicExistenceChecker configures a TopicExistenceChecker using the
// provided Clientset.
func NewRiffTopicExistenceChecker(clientSet *versioned.Clientset) TopicExistenceChecker {
	return &riffTopicExistenceChecker{client: clientSet}
}

func (tec *alwaysTrueTopicExistenceChecker) TopicExists(namespace string, topicName string) (bool, error) {
	return true, nil
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
