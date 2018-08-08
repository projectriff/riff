package service

import (
	"sort"

	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// SelectorMatchFilter() removes all objects that are not services whose
// selector matches the given target.
func SelectorMatchFilter(target map[string]string) filter.ComparableFilter {
	return &serviceForFilter{target}
}

type serviceForFilter struct {
	target map[string]string
}

// Accept() returns true if the object is a Service whose
// selector matches the target fields of the filter.
func (f *serviceForFilter) Accept(obj metav1.Object) bool {
	svc, ok := obj.(*v1.Service)

	if !ok {
		return false
	}

	if len(svc.Spec.Selector) == 0 || len(f.target) == 0 {
		return false
	}

	for k, v := range svc.Spec.Selector {
		if val, ok := f.target[k]; !ok || val != v {
			return false
		}
	}

	return true
}

func (f *serviceForFilter) Equals(other filter.Filter) bool {
	if other, ok := other.(*serviceForFilter); ok {
		return labels.Equals(f.target, other.target)
	}
	return false
}

func PodsFilter(services ...*v1.Service) filter.ComparableFilter {

	// make a copy and sort
	svcs := make([]*v1.Service, len(services))
	copy(svcs, services)

	sort.Slice(svcs, func(i, j int) bool {
		if svcs[i].Namespace != svcs[j].Namespace {
			return svcs[i].Namespace < svcs[j].Namespace
		}
		return svcs[i].Name < svcs[j].Name
	})

	var filters []filter.Filter

	for _, svc := range svcs {
		if len(svc.Spec.Selector) > 0 {
			sfilter := filter.Labels(svc.Spec.Selector)
			nsfilter := filter.NSName(nsname.New(svc.GetNamespace(), ""))
			filters = append(filters, filter.And(nsfilter, sfilter))
		}
	}

	return filter.Or(filters...)
}
