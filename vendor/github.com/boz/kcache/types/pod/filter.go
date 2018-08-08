package pod

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/boz/kcache/filter"
	"k8s.io/api/core/v1"
)

func NodeFilter(names ...string) filter.ComparableFilter {
	set := make(map[string]interface{})
	for _, name := range names {
		set[name] = struct{}{}
	}
	return nodeFilter(set)
}

type nodeFilter map[string]interface{}

func (f nodeFilter) Accept(obj metav1.Object) bool {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return false
	}
	_, ok = f[pod.Spec.NodeName]
	return ok
}

func (f nodeFilter) Equals(other filter.Filter) bool {
	if other, ok := other.(nodeFilter); ok {
		return reflect.DeepEqual(f, other)
	}
	return false
}
