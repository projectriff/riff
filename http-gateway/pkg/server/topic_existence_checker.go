package server

import (
	"github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"

	"log"

	"k8s.io/client-go/tools/cache"

	"fmt"
	"time"

	informers "github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions"
	"github.com/projectriff/riff/kubernetes-crds/pkg/client/informers/externalversions/projectriff/v1alpha1"
	"sync"
)

const (
	defaultNamespace = "default" // expected to be used by consumers of TopicExists
)

// TopicExistenceChecker allows the http-gateway to check for the existence of
// a Topic before attempting to send a message to that topic.
type TopicExistenceChecker interface {
	TopicExists(namespace string, topicName string) bool
}

type riffTopicExistenceChecker struct {
	mutex         *sync.Mutex
	topicInformer v1alpha1.TopicInformer
	knownTopics   map[string]ignoredValue
}

type ignoredValue struct{}

type alwaysTrueTopicExistenceChecker struct{}

// NewAlwaysTrueTopicExistenceChecker configures a TopicExistenceChecker that always returns true.
func NewAlwaysTrueTopicExistenceChecker() TopicExistenceChecker {
	return &alwaysTrueTopicExistenceChecker{}
}

// NewRiffTopicExistenceChecker configures a TopicExistenceChecker using the
// provided Clientset.
func NewRiffTopicExistenceChecker(clientSet *versioned.Clientset) TopicExistenceChecker {
	mutex := &sync.Mutex{}

	riffInformerFactory := informers.NewSharedInformerFactory(clientSet, time.Second*30)
	topicInformer := riffInformerFactory.Projectriff().V1alpha1().Topics()

	knownTopics := make(map[string]ignoredValue)

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

			knownTopics[key] = ignoredValue{}
			log.Printf("New topic has been added: %s", key)
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

			delete(knownTopics, key)
			log.Printf("A topic was removed: %s", key)
		},
	})

	done := make(chan struct{}) //TODO: ideally, this would be the same channel used by the gateway itself
	go topicInformer.Informer().Run(done)

	return &riffTopicExistenceChecker{topicInformer: topicInformer, knownTopics: knownTopics, mutex: mutex}
}

func (tec *alwaysTrueTopicExistenceChecker) TopicExists(namespace string, topicName string) bool {
	return true
}

// TopicExists checks to see if Kubernetes is aware of a riff Topic in a namespace.
func (tec *riffTopicExistenceChecker) TopicExists(namespace string, topicName string) bool {
	tec.mutex.Lock()
	defer tec.mutex.Unlock()

	topicKey := fmt.Sprintf("%s/%s", namespace, topicName)

	_, exists := tec.knownTopics[topicKey]

	return exists
}
