package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Represents the topics.extensions.sk8s.io CRD
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

// Returned in list operations
type TopicList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []Topic `json:"items"`
}
