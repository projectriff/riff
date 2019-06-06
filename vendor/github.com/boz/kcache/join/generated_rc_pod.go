/*
 * AUTO GENERATED - DO NOT EDIT BY HAND
 */

package join

import (
	"context"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/types/pod"
	"github.com/boz/kcache/types/replicationcontroller"
	corev1 "k8s.io/api/core/v1"
)

func RCPodsWith(ctx context.Context,
	srcController replicationcontroller.Controller,
	dstController pod.Publisher,
	filterFn func(...*corev1.ReplicationController) filter.ComparableFilter) (pod.Controller, error) {

	log := logutil.FromContextOrDefault(ctx)

	dst, err := dstController.CloneForFilter()
	if err != nil {
		return nil, err
	}

	update := func(_ *corev1.ReplicationController) {
		objs, err := srcController.Cache().List()
		if err != nil {
			log.Err(err, "join(replicationcontroller,pod: cache list")
			return
		}
		dst.Refilter(filterFn(objs...))
	}

	handler := replicationcontroller.BuildHandler().
		OnInitialize(func(objs []*corev1.ReplicationController) { dst.Refilter(filterFn(objs...)) }).
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
