# RFC-#: A pod with utilities for better developer experience

**Authors:**

**Status:**

**Pull Request URL:** link

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** https://github.com/projectriff/riff/pull/1359 and 


## Problem
The function developer would like to invoke the function for which they just created the `core deployer` or the `knative deployer`.
When the streaming runtime is installed, they would like to test the function by getting events into the stream and look into the contents of the stream. There is no one step mechanism for accomplishing any of these currently.
An [earlier proposal](https://github.com/projectriff/riff/pull/1359) was to provide some of this functionality in the riff cli, this proposal expands on Mark's comment https://github.com/projectriff/riff/pull/1359#discussion_r348617981

### Anti-Goals
We will only address this problem for development/demos, not production, so topics like auth/authz are out of scope for this document.

## Solution
We will ask the developers to run a `riff-developer-utils` pod in their development cluster. This pod will bundle a few binaries/scripts that users can invoke using `kubectl exec`. Since the commands will be run in-cluster, users won't need to run `kubectl port-forward`.
We will have the following commands to start with:
1. **invoke-core:** To invoke the given core deployer.  
The command takes the form:  
    ```
    invoke-core <deployer-name> -n <namespace> <curl-params>
    ```
    where `<curl-params>` are passed along to the post-request.
1. **invoke-knative:** To invoke the given knative deployer.
The command takes the form:  
    ```
    invoke-knative <deployer-name> -n <namespace> <curl-params>
    ```
    where `<curl-params>` are passed along to the post-request.

1. **publish-stream:** To publish an event to the given stream.
The command takes the form:
    ```
    publish-stream <stream-name> -n <namespace> --payload <payload-as-string> --content-type <content-type> --header <headers-as-string>
    ```
    where `stream-name`, `payload` and `content-type` are mandatory and `headers` can be used multiple times.
1. **subscribe-stream:** To subscribe for events from the given stream.
The command takes the form:
    ```
    riff streaming stream subscribe <stream-name> --offset <long-offset> --decode-payload
    ```
    This will display all the events in the stream from the given offset in the following json format:
    ```
    {
        "payload": "base64 encoded user payload",
        "content-type": "the content type of the message",
        "headers": {"user provided header": "while publishing"}
    }
    ```
    The payload will be base64 encoded by default, however if the `decode-payload` flag is specified, then it will be decoded as string.

The namespace parameter is optional for all the commands. If not specified, namespace of the `riff-developer-utils` pod will be assumed.

### User Impact
The user will be able to able to invoke functions/deployers and publish and subscribe to streams using the following workflow:
1. Run the `riff-developer-utils` pod in their cluster with something like:
```
kubectl run --image=projectriff/developer-utils:latest
```
1. Invoke the commands specified above using kubectl exec. An example follows:
```
kubectl exec -it riff-developer-utils -- invoke-core upper -d test
```


### Backwards Compatibility and Upgrade Path
*How do the proposed changes impact backwards-compatibility? Are APIs and CLI commands changing?*

*Is there a need for a deprecation process to provide an upgrade path to users?*

## FAQ
*Answers to common questions that may arise and those that you’ve commonly been asked after requesting comments for this proposal.*
