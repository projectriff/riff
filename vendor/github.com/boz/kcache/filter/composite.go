package filter

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type andFilter []Filter

func And(children ...Filter) ComparableFilter {
	return andFilter(children)
}

func (f andFilter) Accept(obj metav1.Object) bool {
	for _, child := range f {
		if !child.Accept(obj) {
			return false
		}
	}
	return true
}

func (f andFilter) Equals(other Filter) bool {
	if other, ok := other.(andFilter); ok {
		return compareFilterList(f, other)
	}
	return false
}

type orFilter []Filter

func Or(children ...Filter) ComparableFilter {
	return orFilter(children)
}

func (f orFilter) Accept(obj metav1.Object) bool {
	for _, child := range f {
		if child.Accept(obj) {
			return true
		}
	}
	return false
}

func (f orFilter) Equals(other Filter) bool {
	if other, ok := other.(orFilter); ok {
		return compareFilterList(f, other)
	}
	return false
}

func compareFilterList(a []Filter, b []Filter) bool {
	if len(a) != len(b) {
		return false
	}

	// must be in same order
	for idx := range a {

		fa, ok := a[idx].(ComparableFilter)
		if !ok {
			return false
		}

		fb, ok := b[idx].(ComparableFilter)
		if !ok {
			return false
		}

		if !fa.Equals(fb) {
			return false
		}
	}

	return true
}
