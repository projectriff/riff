package replicationcontroller

import (
	"sort"

	"github.com/boz/kcache/filter"
	corev1 "k8s.io/api/core/v1"
)

func PodsFilter(sources ...*corev1.ReplicationController) filter.ComparableFilter {

	// make a copy and sort
	srcs := make([]*corev1.ReplicationController, len(sources))
	copy(srcs, sources)

	sort.Slice(srcs, func(i, j int) bool {
		if srcs[i].Namespace != srcs[j].Namespace {
			return srcs[i].Namespace < srcs[j].Namespace
		}
		return srcs[i].Name < srcs[j].Name
	})

	filters := make([]filter.Filter, 0, len(srcs))

	for _, svc := range srcs {
		filters = append(filters, filter.Labels(svc.Spec.Selector))
	}

	return filter.Or(filters...)
}
