/*
 * AUTO GENERATED - DO NOT EDIT BY HAND
 */

package join

import (
	"context"

	"k8s.io/api/extensions/v1beta1"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/types/pod"
	"github.com/boz/kcache/types/replicaset"
)

func RSPodsWith(ctx context.Context,
	srcController replicaset.Controller,
	dstController pod.Publisher,
	filterFn func(...*v1beta1.ReplicaSet) filter.ComparableFilter) (pod.Controller, error) {

	log := logutil.FromContextOrDefault(ctx)

	dst, err := dstController.CloneForFilter()
	if err != nil {
		return nil, err
	}

	update := func(_ *v1beta1.ReplicaSet) {
		objs, err := srcController.Cache().List()
		if err != nil {
			log.Err(err, "join(replicaset,pod: cache list")
			return
		}
		dst.Refilter(filterFn(objs...))
	}

	handler := replicaset.BuildHandler().
		OnInitialize(func(objs []*v1beta1.ReplicaSet) { dst.Refilter(filterFn(objs...)) }).
		OnCreate(update).
		OnUpdate(update).
		OnDelete(update).
		Create()

	monitor, err := replicaset.NewMonitor(srcController, handler)
	if err != nil {
		dst.Close()
		return nil, log.Err(err, "join(replicaset,pod): monitor")
	}

	go func() {
		<-dst.Done()
		monitor.Close()
	}()

	return dst, nil
}
