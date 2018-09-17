package join

import (
	"context"

	"github.com/boz/kcache/types/daemonset"
	"github.com/boz/kcache/types/deployment"
	"github.com/boz/kcache/types/ingress"
	"github.com/boz/kcache/types/pod"
	"github.com/boz/kcache/types/replicaset"
	"github.com/boz/kcache/types/replicationcontroller"
	"github.com/boz/kcache/types/service"
)

func ServicePods(ctx context.Context,
	src service.Controller, dst pod.Publisher) (pod.Controller, error) {
	return ServicePodsWith(ctx, src, dst, service.PodsFilter)
}

func RCPods(ctx context.Context,
	src replicationcontroller.Controller, dst pod.Publisher) (pod.Controller, error) {
	return RCPodsWith(ctx, src, dst, replicationcontroller.PodsFilter)
}

func RSPods(ctx context.Context,
	src replicaset.Controller, dst pod.Publisher) (pod.Controller, error) {
	return RSPodsWith(ctx, src, dst, replicaset.PodsFilter)
}

func DeploymentPods(ctx context.Context,
	src deployment.Controller, dst pod.Publisher) (pod.Controller, error) {
	return DeploymentPodsWith(ctx, src, dst, deployment.PodsFilter)
}

func DaemonSetPods(ctx context.Context,
	src daemonset.Controller, dst pod.Publisher) (pod.Controller, error) {
	return DaemonSetPodsWith(ctx, src, dst, daemonset.PodsFilter)
}

func IngressServices(ctx context.Context,
	src ingress.Controller, dst service.Publisher) (service.Controller, error) {
	return IngressServicesWith(ctx, src, dst, ingress.ServicesFilter)
}

func IngressPods(ctx context.Context, srcbase ingress.Controller, svcbase service.Controller, dstbase pod.Controller) (pod.Controller, error) {
	svcs, err := IngressServices(ctx, srcbase, svcbase)
	if err != nil {
		return nil, err
	}

	pods, err := ServicePods(ctx, svcs, dstbase)
	if err != nil {
		svcs.Close()
		return nil, err
	}
	return pods, nil
}
