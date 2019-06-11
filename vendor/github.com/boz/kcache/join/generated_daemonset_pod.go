/*
 * AUTO GENERATED - DO NOT EDIT BY HAND
 */

package join

import (
	"context"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/types/daemonset"
	"github.com/boz/kcache/types/pod"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
)

func DaemonSetPodsWith(ctx context.Context,
	srcController daemonset.Controller,
	dstController pod.Publisher,
	filterFn func(...*extv1beta1.DaemonSet) filter.ComparableFilter) (pod.Controller, error) {

	log := logutil.FromContextOrDefault(ctx)

	dst, err := dstController.CloneForFilter()
	if err != nil {
		return nil, err
	}

	update := func(_ *extv1beta1.DaemonSet) {
		objs, err := srcController.Cache().List()
		if err != nil {
			log.Err(err, "join(daemonset,pod: cache list")
			return
		}
		dst.Refilter(filterFn(objs...))
	}

	handler := daemonset.BuildHandler().
		OnInitialize(func(objs []*extv1beta1.DaemonSet) { dst.Refilter(filterFn(objs...)) }).
		OnCreate(update).
		OnUpdate(update).
		OnDelete(update).
		Create()

	monitor, err := daemonset.NewMonitor(srcController, handler)
	if err != nil {
		dst.Close()
		return nil, log.Err(err, "join(daemonset,pod): monitor")
	}

	go func() {
		<-dst.Done()
		monitor.Close()
	}()

	return dst, nil
}
