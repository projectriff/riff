package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TopicPlural      string = "topics"
	TopicGroup       string = "projectriff.io"
	TopicVersion     string = "v1"
	FullTopicCRDName    string = TopicPlural + "." + TopicGroup
)

var defaultPartitions = int32(1)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

// Represents the topics.projectriff.io CRD
type Topic struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               TopicSpec   `json:"spec"`
	Status             TopicStatus `json:"status,omitempty"`
}

// Spec (what the user wants) for a topic
type TopicSpec struct {

	// TODO: add fields here. Java had name (if != from metadata.name?), partitions, exposeRead/exposeWrite

	// +optional
	Partitions *int32 `json:"partitions,omitempty"`
}

// Status (computed) for a topic
type TopicStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Returned in list operations
type TopicList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Topic `json:"items"`
}

func SetDefaults_TopicSpec(obj *TopicSpec) {
	if obj.Partitions == nil {
		obj.Partitions = &defaultPartitions
	}
}