package server

import (
	"github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"

	"log"

	"k8s.io/client-go/tools/cache"

	"fmt"
	"time"

	informers "github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions"
	"github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions/projectriff.io/v1alpha1"
	"sync"
)

// TopicExistenceChecker allows the http-gateway to check for the existence of
// a Topic before attempting to send a message to that topic.
type TopicExistenceChecker interface {
	TopicExists(topicName string) bool
}

type riffTopicExistenceChecker struct {
	topicInformer v1alpha1.TopicInformer

	mutex       *sync.Mutex
	knownTopics map[string]map[string]ignoredValue
}

type ignoredValue struct{}

type alwaysTrueTopicExistenceChecker struct{}

// NewAlwaysTrueTopicExistenceChecker configures a TopicExistenceChecker that always returns true.
func NewAlwaysTrueTopicExistenceChecker() TopicExistenceChecker {
	return &alwaysTrueTopicExistenceChecker{}
}

// NewRiffTopicExistenceChecker configures a TopicExistenceChecker using the provided Clientset.
func NewRiffTopicExistenceChecker(clientSet *versioned.Clientset, stop <-chan struct{}) TopicExistenceChecker {
	riffInformerFactory := informers.NewSharedInformerFactory(clientSet, time.Second*30)
	topicInformer := riffInformerFactory.Projectriff().V1alpha1().Topics()

	mutex := &sync.Mutex{}
	knownTopics := make(map[string]map[string]ignoredValue)

	topicInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mutex.Lock()
			defer mutex.Unlock()

			//TODO: implement riff-specific KeyFunc https://github.com/projectriff/riff/pull/558#discussion_r184437224
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				// It is likely that the key is faulty, but we cannot signal an error.
				// To prevent errors for processing a bad key, we return after logging.
				log.Printf("AddFunc failed during a topic lookup: %#v", err)
				return
			}

			namespace, topic := splitTopicNamespace(key)
			knownNamespaces := knownTopics[topic]
			if knownNamespaces == nil {
				knownNamespaces = make(map[string]ignoredValue)
				log.Printf("Topic has been added: %s", topic)
			} else {
				log.Printf("Warning, a duplicate topic has been added: %s in namespace %s", topic, namespace)
			}
			knownNamespaces[namespace] = ignoredValue{}
			knownTopics[topic] = knownNamespaces
		},

		DeleteFunc: func(obj interface{}) {
			mutex.Lock()
			defer mutex.Unlock()

			//TODO: implement riff-specific KeyFunc https://github.com/projectriff/riff/pull/558#discussion_r184437224
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				// It is likely that the key is faulty, but we cannot signal an error.
				// To prevent errors for processing a bad key, we return after logging.
				log.Printf("DeleteFunc failed during a topic lookup: %#v", err)
				return
			}

			namespace, topic := splitTopicNamespace(key)
			knownNamespaces := knownTopics[topic]
			if knownNamespaces != nil {
				delete(knownNamespaces, namespace)
			}
			if len(knownNamespaces) == 0 {
				delete(knownTopics, topic)
				log.Printf("Topic has been removed: %s", topic)
			} else {
				log.Printf("Duplicate topic has been removed: %s in namespace %s", topic, namespace)
			}
		},
	})

	go topicInformer.Informer().Run(stop)

	return &riffTopicExistenceChecker{topicInformer: topicInformer, mutex: mutex, knownTopics: knownTopics}
}

func splitTopicNamespace(key string) (string, string) {
	namespace, topic, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("SplitMetaNamespaceKey failed during a topic lookup: %#v", err)
		return "", key
	}
	return namespace, topic
}

func (tec *alwaysTrueTopicExistenceChecker) TopicExists(topicName string) bool {
	return true
}

// TopicExists checks to see if the gateway is aware of a riff Topic in any namespace.
func (tec *riffTopicExistenceChecker) TopicExists(topicName string) bool {
	tec.mutex.Lock()
	defer tec.mutex.Unlock()

	topicKey := fmt.Sprintf("%s", topicName)

	_, exists := tec.knownTopics[topicKey]

	return exists
}
