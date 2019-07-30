---
title: "Announcing riff v0.1.1 on Knative"
---

We are pleased to announce that [riff v0.1.1](https://github.com/projectriff/riff/releases/tag/v0.1.1) on Knative is now available. Thanks, everyone for contributing.

<!--truncate-->

## Uninstall

If you are using the latest riff on Knative for demos, you might find it useful to do a quick uninstall without zapping your whole cluster and cached images.

```sh
# remove everything including istio without prompting 
riff system uninstall --istio --force
```

Istio will remain installed if you omit `--istio`. 

## Command function invoker

With this release the command invoker is back in business.  This invoker accepts HTTP requests and invokes a command for each request. For example, let's try a simple shell script

```sh
#!/bin/sh

xargs echo -n hello
```

Create the function using the riff CLI. This assumes that you are running riff with image push credentials to DockerHub with YOUR-DOCKER-ID.

```sh
riff function create command hello \
  --git-repo https://github.com/markfisher/riff-sample-hello.git \
  --artifact hello.sh \
  --image YOUR-DOCKER-ID/demo-command-hello
```

Once the function has been built, you should be able to run it using

```sh
riff service invoke hello -- -w '\n' -d stranger
```

The output should look something like this:

```
curl http://192.168.64.31:32380 -H 'Host: hello.default.example.com' -w '\n' -d stranger
hello stranger
```
