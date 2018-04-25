package server

import (
	"github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"

	"k8s.io/client-go/tools/cache"
	"log"

	informers "github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions"
	"time"
	"github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions/projectriff/v1alpha1"
	"fmt"
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
	topicInformer v1alpha1.TopicInformer
	knownTopics map[string]ignoredValue
}

type ignoredValue struct {
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
	riffInformerFactory := informers.NewSharedInformerFactory(clientSet, time.Second*30)
	topicInformer := riffInformerFactory.Projectriff().V1alpha1().Topics()

	knownTopics := make(map[string]ignoredValue)

	topicInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				knownTopics[key] = ignoredValue{}
				log.Printf("Added topic to internal map: %+v", key)
			}


		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				delete(knownTopics, key)
				log.Printf("Removed topic from internal map: %+v", key)
			}
		},
	})

	done := make(chan struct{}) //TODO: ideally, this would be the same channel used by the gateway itself
	go topicInformer.Informer().Run(done)

	return &riffTopicExistenceChecker{topicInformer: topicInformer, knownTopics: knownTopics}
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
	topicKey := fmt.Sprintf("%s/%s", namespace, topicName)

	if _, exists := tec.knownTopics[topicKey]; exists {
		return true, nil
	} else {
		return false, nil
	}
}
