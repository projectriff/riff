package filter

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Labels() returns a filter which returns true if
// the provided map is a subset of the object's labels.
func Labels(match map[string]string) ComparableFilter {
	return Selector(labels.SelectorFromSet(match))
}

func LabelSelector(ls *metav1.LabelSelector) ComparableFilter {
	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		// todo: return error
		panic("invalid selector")
	}
	return Selector(selector)
}

func Selector(selector labels.Selector) ComparableFilter {
	// assumes selector is sorted
	return &selectorFilter{selector}
}

type selectorFilter struct {
	selector labels.Selector
}

func (f *selectorFilter) Accept(obj metav1.Object) bool {
	return f.selector.Matches(labels.Set(obj.GetLabels()))
}

func (f *selectorFilter) Equals(other Filter) bool {
	if other, ok := other.(*selectorFilter); ok {
		return reflect.DeepEqual(f.selector, other.selector)
	}
	return false
}
