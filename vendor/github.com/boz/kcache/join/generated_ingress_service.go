/*
 * AUTO GENERATED - DO NOT EDIT BY HAND
 */

package join

import (
	"context"

	"k8s.io/api/extensions/v1beta1"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/types/ingress"
	"github.com/boz/kcache/types/service"
)

func IngressServicesWith(ctx context.Context,
	srcController ingress.Controller,
	dstController service.Publisher,
	filterFn func(...*v1beta1.Ingress) filter.ComparableFilter) (service.Controller, error) {

	log := logutil.FromContextOrDefault(ctx)

	dst, err := dstController.CloneForFilter()
	if err != nil {
		return nil, err
	}

	update := func(_ *v1beta1.Ingress) {
		objs, err := srcController.Cache().List()
		if err != nil {
			log.Err(err, "join(ingress,service: cache list")
			return
		}
		dst.Refilter(filterFn(objs...))
	}

	handler := ingress.BuildHandler().
		OnInitialize(func(objs []*v1beta1.Ingress) { dst.Refilter(filterFn(objs...)) }).
		OnCreate(update).
		OnUpdate(update).
		OnDelete(update).
		Create()

	monitor, err := ingress.NewMonitor(srcController, handler)
	if err != nil {
		dst.Close()
		return nil, log.Err(err, "join(ingress,service): monitor")
	}

	go func() {
		<-dst.Done()
		monitor.Close()
	}()

	return dst, nil
}
