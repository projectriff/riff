---
id: riff-handler
title: "riff handler"
---
## riff handler

handlers map HTTP requests to applications, functions or images

### Synopsis

Handlers can be created for one of an application, function or image.
Application and function based handlers continuously watch for the latest built
image and will deploy new images. If the underlying application or function is
deleted, the handler will continue to run, but will no longer self update. Image
based handlers must be manually updated to trigger roll out of an updated image.

Applications, functions and images are logically equivalent at runtime.
Functions with an invoker are more focused and opinionated applications, and
images are compiled applications.

Users wishing to perform checks on built images before deploying them can
provide their own external process to watch the application/function for new
images and only update the handler image once those checks pass.

The hostname to access the handler is available in the handler listing.

### Options

```
  -h, --help   help for handler
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions
* [riff handler create](riff_handler_create.md)	 - create a handler to map HTTP requests to an application, function or image
* [riff handler delete](riff_handler_delete.md)	 - delete handler(s)
* [riff handler list](riff_handler_list.md)	 - table listing of handlers
* [riff handler status](riff_handler_status.md)	 - show handler status
* [riff handler tail](riff_handler_tail.md)	 - watch handler logs

