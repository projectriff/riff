## riff function invoke

Invoke a function

### Synopsis

Invoke a function by shelling out to curl.

The curl command is printed so it can be copied and extended.

Additional curl arguments and flags may be specified after a double dash (--).

```
riff function invoke [flags]
```

### Examples

```
  riff function invoke square --namespace joseph-ns
  riff function invoke square /foo -- --data 42
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

* [riff function](riff_function.md)	 - Interact with function related resources

