/*
 * AUTO GENERATED - DO NOT EDIT BY HAND
 */

package join

import (
	"context"

	"k8s.io/api/core/v1"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/types/pod"
	"github.com/boz/kcache/types/replicationcontroller"
)

func RCPodsWith(ctx context.Context,
	srcController replicationcontroller.Controller,
	dstController pod.Publisher,
	filterFn func(...*v1.ReplicationController) filter.ComparableFilter) (pod.Controller, error) {

	log := logutil.FromContextOrDefault(ctx)

	dst, err := dstController.CloneForFilter()
	if err != nil {
		return nil, err
	}

	update := func(_ *v1.ReplicationController) {
		objs, err := srcController.Cache().List()
		if err != nil {
			log.Err(err, "join(replicationcontroller,pod: cache list")
			return
		}
		dst.Refilter(filterFn(objs...))
	}

	handler := replicationcontroller.BuildHandler().
		OnInitialize(func(objs []*v1.ReplicationController) { dst.Refilter(filterFn(objs...)) }).
		OnCreate(update).
		OnUpdate(update).
		OnDelete(update).
		Create()

	monitor, err := replicationcontroller.NewMonitor(srcController, handler)
	if err != nil {
		dst.Close()
		return nil, log.Err(err, "join(replicationcontroller,pod): monitor")
	}

	go func() {
		<-dst.Done()
		monitor.Close()
	}()

	return dst, nil
}
