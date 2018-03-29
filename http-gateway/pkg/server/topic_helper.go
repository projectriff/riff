package server

import (
	"log"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"
	riffcs "github.com/projectriff/riff/kubernetes-crds/pkg/client/clientset/versioned"

	"k8s.io/client-go/rest"
)

type TopicHelper interface {
	TopicExists(topicName string) bool
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
		glog.Fatalf("Error building riff clientset: %s", err.Error())
	}

	return &topicHelper{client: riffClient}, nil

}

func (tw *topicHelper) TopicExists(topicName string) bool {
	_, err := tw.client.ProjectriffV1alpha1().Topics("default").Get(topicName, v1.GetOptions{})

	if err != nil {
		log.Printf("%s\n", err)
		return false
	}

	return true
}
