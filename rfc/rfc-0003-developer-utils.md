# RFC-0003: A pod with utilities for better developer experience

**Authors:** Swapnil Bawaskar

**Status:**

**Pull Request URL:** [#1362](https://github.com/projectriff/riff/pull/1362)

**Superseded by:** N/A

**Supersedes:** [RFC 0001](https://github.com/projectriff/riff/pull/1359)

**Related:**


## Problem
A developer would like to interact with riff resources inside the cluster in ways that may not be desirable to expose outside of the cluster. This may be reading or writing message from a stream or invoking an http workload whose ingress policy is cluster-local. There is no one step mechanism for accomplishing any of these currently.

An [earlier proposal](https://github.com/projectriff/riff/pull/1359) was to provide some of this functionality in the riff cli. The current proposal is a response to [feedback](https://github.com/projectriff/riff/pull/1359#discussion_r348617981) on that earlier proposal.

### Anti-Goals
We will only address this problem for development/demos, not production, so topics like auth/authz are out of scope for this document.

## Solution
We will ask the developers to run a `riff-dev` pod in their development cluster. This pod will bundle a few commands that users can invoke using `kubectl exec`. Since the commands will be run in-cluster, users won't need to run `kubectl port-forward`, however, the pod will need to be run with a service account that has appropriate RBAC permissions.

We will have the following commands to start with:

1. **publish:** To publish an event to the given stream.
    
    The command takes the form:
    
    ```
    publish <stream-name> --content-type <content-type> [--payload <payload-as-plain-text>] [--payload-base64 <payload-as-base64-text>] [--header "<header-name>: <header-value>"]
    ```
    
    where `stream-name`, and `--content-type` are mandatory and `--header` can be used multiple times. `--payload` and `--payload-base64` are mutually exclusive. Binary data should only be passed via `--payload-base64` after it has been `base64` encoded.

1. **subscribe:** To subscribe for events from the given stream.
    The command takes the form:
    
    ```
    subscribe <stream-name> [--from-beginning]
    ```
    
    If the `--from-beginning` option is present, display all the events in the stream, otherwise only new events are displayed as [JSON Lines](http://jsonlines.org) in the following form:
    
    ```
    {"payload":"base64 encoded message payload","content-type":"the content type of the message","headers":{"header name": "header value"}}
    {"payload":"base64 encoded message payload","content-type":"the content type of the message","headers":{"header name": "header value"}}
    ```
    
    Since the payload may contain binary data, it will be displayed as a base64 encoded string. Tools like `jq` and `base64 --decode` are useful to get a raw payload.
    
    The subscribe command will run until terminated (SIGTERM), or an error occurs on the stream.

1. [**jq**](https://stedolan.github.io/jq/): Process JSON.

1. [**base64:**](http://manpages.ubuntu.com/manpages/bionic/man1/base64.1.html) Encode and decode base64 strings.

1. [**curl:**](https://curl.haxx.se) Make HTTP requests.

Each command targets resources in the same namespace as the pod is running. Additional commands and behaviors may be defined by future RFCs.

### User Impact
The user will be able to run these using the following workflow:

1. Run the `riff-dev` pod in their cluster with something like:
    
    ```
    # the pod should be run with a service account that has appropriate access to the api server
    kubectl run riff-dev --image=<image>
    ```

1. Invoke the commands specified above using kubectl exec. An example follows:
    
    ```
    kubectl exec -it riff-dev -- <command to run inside pod with args>
    ```

The implementation should recommend specific RBAC policies needed to use the commands in the pod as well as specific commands and arguments.

### Backwards Compatibility and Upgrade Path
Net new functionality, no impact on backward compatibility.

## FAQ
