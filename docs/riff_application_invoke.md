## riff application invoke

Invoke a application

### Synopsis

Invoke a application by shelling out to curl.

The curl command is printed so it can be copied and extended.

Additional curl arguments and flags may be specified after a double dash (--).

```
riff application invoke [flags]
```

### Examples

```
  riff application invoke square --namespace joseph-ns
  riff application invoke square /foo -- --data 42
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

* [riff application](riff_application.md)	 - Interact with application related resources

