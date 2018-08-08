package replicationcontroller

import (
	"sort"

	"github.com/boz/kcache/filter"
	"k8s.io/api/core/v1"
)

func PodsFilter(sources ...*v1.ReplicationController) filter.ComparableFilter {

	// make a copy and sort
	srcs := make([]*v1.ReplicationController, len(sources))
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
