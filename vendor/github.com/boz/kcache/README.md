# kcache: kubernetes object cache [![Build Status](https://travis-ci.org/boz/kcache.svg?branch=master)](https://travis-ci.org/boz/kcache) [![codecov](https://codecov.io/gh/boz/kcache/branch/master/graph/badge.svg)](https://codecov.io/gh/boz/kcache) 

Kcache is a [kubernetes](https://github.com/kubernetes/kubernetes) object data source similar to [k8s.io/client-go/tools/cache](https://github.com/kubernetes/client-go/tree/master/tools/cache) which uses channels to create a flexible event-based toolkit.  Features include [typed producers](#types), [joining between multiple producers](#joins), and [(re)filtering](#filtering).

 * [Usage](#usage)
   * [Controllers](#controllers)
   * [Channels](#channels)
   * [Callbacks](#callbacks)
   * [Types](#types)
   * [Joins](#joins)
   * [Filtering](#filters)

Kcache was originally created to drive a Kubernetes monitoring application and it currently powers [kail](https://github.com/boz/kail).

## Usage

Using kcache involves creating [controllers](#controllers) to manage dynamic object sets with the kubernetes API.  The monitored objects are cached and events about changing state are broadcast to subscribers.

### Controllers

Each controller represents a single kubernetes watch stream.  There can be any number of subscribers to
each controller, and subscribers can be publishers themselves.

```go
  controller, err := kcache.NewController(ctx,log,client)

  // wait for the initial sync to be complete
  <-controller.Ready()

  fmt.Println("controller has been synced")
```

Controllers maintain a cache of the objects being watched.

```go
  // fetch the pod named 'pod-1' in the namespace 'default' from the cache.
  pod, err := controller.Cache().Get("default","pod-1")
```

### Channels

There are many ways to subscribe to a controller's events, the most basic is a simple channel-based subscription:

```go
  sub, err := controller.Subscribe()
  <-sub.Ready()

  // fetch cached list of objects
  sub.Cache().List()

  for event := range sub.Events() {
    // handle add/update/delete event for objects
  }
```

### Callbacks

In addition to [channels](#channels), callbacks can be used to handle events

```go
  handler := kcache.BuildHandler().
    OnInitialize(func(objs []metav1.Object) { /* ... */ }).
    OnCreate(func(obj metav1.Object){ /* ... */ }).
    OnUpdate(func(obj metav1.Object){ /* ... */ }).
    OnDelete(func(obj metav1.Object){ /* ... */ }).
    Create()
  controller

  kcache.NewMonitor(controller,handler)
```

### Types

Typed controllers and subscribers are available to reduce the need for casting objects.  Each type has all of the features of the untyped system (channels,callbacks, filtering, caches, etc...)

```go
  controller, err := pod.NewController(ctx,log,client,"default")
  sub, err := controller.Subescribe()
  ...
```

Currently implemented types are:
 
 * Pod
 * Node
 * Event
 * Secret
 * Service
 * Ingress
 * Daemonset
 * ReplicaSet
 * Deployment
 * ReplicationController

### Filtering

The cache and events that are be exposed to a subscription can be limited by a filter object

The following will return a subscription that only sees the pod named "default/pod-1" pod in its cache and events:

```go
  sub, err := controller.SubscribeWithFilter(filter.NSName("default","pod-1"))
```

Additionally, new publishers can be created with filters.  In the following example,
`sub_a` will only receive events about "default/pod-1" and `sub_b` will only receive events about "default/pod-2"

```go
  pub_a, err := controller.CloneWithFilter(filter.NSName("default","pod-1"))
  pub_b, err := controller.CloneWithFilter(filter.NSName("default","pod-2"))

  sub_a, err := pub_a.Subscribe()
  sub_b, err := pub_b.Subscribe()
```

### Refiltering

The filter used for filtered publishers and subscribers can be changed at any time.  The cache for each will readjust and `CREATE`, `DELETE` events will be emitted as necessary.

In the example below, if the pods "default/pod-1" and "default/pod-2" exist, `sub_a` will receive a delete event for "default/pod-1" and a create event for "default/pod-2"

```go
  pub_a, err := controller.CloneWithFilter(filter.NSName("default","pod-1"))

  sub_a, err := pub_a.Subscribe()

  <-sub_a.Ready()

  go func() {
    for evt := sub_a.Events() {
      fmt.Println(evt)
    }
  }()

  pub_a.Refilter(filter.NSName("default","pod-2"))
```

### Joins

[Refiltering](#refiltering) allows for joining between different publishers.  The join is dynamic -- as the objects of the joined
publisher changes, so does the set of objects in the resulting publisher.

In the example below, `sub` will only know about pods that are targeted by the "default/frontend" service.

```go
  pods, err := pod.NewController(/*...*/)
  services, err := service.NewController(/*...*/)

  frontend, err := services.CloneWithFilter(filter.NSName("default","frontend"))

  sub, err := join.ServicePods(ctx,frontend,pods)

  <- sub.Ready()

  for evt := range sub.Events() {
    /* ... */
  }
```

Joining can be done by hand but there are a number of utility joins available:

 * `ServicePods()` - restrict pods to those that match the services available in the given publisher.
 * `RCPods()` - restrict pods to those that match the replication controllers in the given publisher.
 * `RSPods()` - restrict pods to those that match the replica sets in the given publisher.
 * `DeploymentPods()` - restrict pods to those that match the deployments in the given publisher.
 * `DaemonSetPods()` - restrict pods to those that match the daemonsets in the given publisher.
 * `IngressServices()` - restrict services to those that match the ingresses in the given publisher.
 * `IngressPods()` - restrict pods to those that match the services which match the ingresses in the given publisher (_double join_)
