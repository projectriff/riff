# RFC-0003: A pod with utilities for better developer experience

**Authors:** Swapnil Bawaskar

**Status:**

**Pull Request URL:** [#1362](https://github.com/projectriff/riff/pull/1362)

**Superseded by:** N/A

**Supersedes:** [RFC 0001](https://github.com/projectriff/riff/pull/1359)

**Related:**


## Problem
A developer would like to invoke the function for which they just created the `core deployer` or the `knative deployer`.
When the streaming runtime is installed, they would like to test the processor by getting events into the stream and look into the contents of the stream. There is no one step mechanism for accomplishing any of these currently.
An [earlier proposal](https://github.com/projectriff/riff/pull/1359) was to provide some of this functionality in the riff cli. The current proposal is a response to [feedback](https://github.com/projectriff/riff/pull/1359#discussion_r348617981) on that earlier proposal.

### Anti-Goals
We will only address this problem for development/demos, not production, so topics like auth/authz are out of scope for this document.

## Solution
We will ask the developers to run a `riff-developer-utils` pod in their development cluster. This pod will bundle a few binaries/scripts that users can invoke using `kubectl exec`. Since the commands will be run in-cluster, users won't need to run `kubectl port-forward`.
We will have the following commands to start with:
1. **invoke-core:** To invoke the given core deployer.  
The command takes the form:  
    ```
    invoke-core <deployer-name> -n <namespace> -- <curl-params>
    ```
    where everything after `--` is passed as parameter to curl
1. **invoke-knative:** To invoke the given knative deployer.
The command takes the form:  
    ```
    invoke-knative <deployer-name> -n <namespace> -- <curl-params>
    ```
    where everything after `--` is passed as parameter to curl

1. **publish:** To publish an event to the given stream.
The command takes the form:
    ```
    publish <stream-name> -n <namespace> --payload <payload-as-string> --content-type <content-type> --header "<header-name>: <header-value>"
    ```
    where `stream-name`, `payload` and `content-type` are mandatory and `header` can be used multiple times.
1. **subscribe:** To subscribe for events from the given stream.
The command takes the form:
    ```
    subscribe <stream-name> --payload-as-string --from-beginning
    ```
    If the `--from-beginning` option is present, display all the events in the stream, otherwise only new events are displayed in the following json format:
    ```
    {"payload": "base64 encoded user payload","content-type": "the content type of the message","headers": {"user provided header": "while publishing"}}
    ```
    The payload will be base64 encoded unless the `--payload-as-string` flag is present, in which case it will be displayed as a string.

The namespace parameter is optional for all the commands. If not specified, the namespace of the `riff-developer-utils` pod will be assumed.

### User Impact
The user will be able to able to invoke functions/deployers and publish and subscribe to streams using the following workflow:
1. Run the `riff-developer-utils` pod in their cluster with something like:
    ```
    kubectl run --image=projectriff/developer-utils
    ```
1. Invoke the commands specified above using kubectl exec. An example follows:
    ```
    kubectl exec -it riff-developer-utils -- invoke-core upper -d test
    ```


### Backwards Compatibility and Upgrade Path
Net new functionality, no impact on backward compatibility.

## FAQ
