---
id: riff-service-invoke
title: "riff service invoke"
---
## riff service invoke

Invoke a service

### Synopsis

Invoke a service by shelling out to curl.

The curl command is printed so it can be copied and extended.

Additional curl arguments and flags may be specified after a double dash (--).

```
riff service invoke [flags]
```

### Examples

```
  riff service invoke square --namespace joseph-ns
  riff service invoke square /foo -- --data 42
```

### Options

```
  -h, --help                  help for invoke
      --json                  set the request's content type to 'application/json'
  -n, --namespace namespace   the namespace of the service
      --text                  set the request's content type to 'text/plain'
```

### Options inherited from parent commands

```
      --kubeconfig path   the path of a kubeconfig (default "~/.kube/config")
      --master address    the address of the Kubernetes API server; overrides any value in kubeconfig
```

### SEE ALSO

* [riff service](riff_service.md)	 - Interact with service related resources

