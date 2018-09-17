package kcache

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EventType string

const (
	EventTypeCreate EventType = "create"
	EventTypeUpdate EventType = "update"
	EventTypeDelete EventType = "delete"
)

type Event interface {
	Type() EventType
	Resource() v1.Object
}

type event struct {
	eventType EventType
	resource  v1.Object
}

func NewEvent(et EventType, resource v1.Object) Event {
	return event{et, resource}
}

func (e event) Type() EventType {
	return e.eventType
}

func (e event) Resource() v1.Object {
	return e.resource
}

func (e event) String() string {
	return fmt.Sprintf(
		"Event{%v %v/%v}", e.eventType, e.Resource().GetNamespace(), e.resource.GetName())
}
