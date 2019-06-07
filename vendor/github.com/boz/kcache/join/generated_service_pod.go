/*
 * AUTO GENERATED - DO NOT EDIT BY HAND
 */

package join

import (
	"context"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/types/pod"
	"github.com/boz/kcache/types/service"
	corev1 "k8s.io/api/core/v1"
)

func ServicePodsWith(ctx context.Context,
	srcController service.Controller,
	dstController pod.Publisher,
	filterFn func(...*corev1.Service) filter.ComparableFilter) (pod.Controller, error) {

	log := logutil.FromContextOrDefault(ctx)

	dst, err := dstController.CloneForFilter()
	if err != nil {
		return nil, err
	}

	update := func(_ *corev1.Service) {
		objs, err := srcController.Cache().List()
		if err != nil {
			log.Err(err, "join(service,pod: cache list")
			return
		}
		dst.Refilter(filterFn(objs...))
	}

	handler := service.BuildHandler().
		OnInitialize(func(objs []*corev1.Service) { dst.Refilter(filterFn(objs...)) }).
		OnCreate(update).
		OnUpdate(update).
		OnDelete(update).
		Create()

	monitor, err := service.NewMonitor(srcController, handler)
	if err != nil {
		dst.Close()
		return nil, log.Err(err, "join(service,pod): monitor")
	}

	go func() {
		<-dst.Done()
		monitor.Close()
	}()

	return dst, nil
}
