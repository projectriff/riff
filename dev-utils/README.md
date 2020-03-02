# Developer Tools
This repository provides tools for riff users to develop and debug functions. These tools are bundled in a container image that is meant to be run in the development cluster.

## Using the tools
These tools can be used by running a pod in the k8s cluster with a configuration like this:
```bash
# create a service account to run with
kubectl create serviceaccount riff-dev --namespace=${NAMESPACE}
# grant that service account edit access to resources in the namespace
kubectl create rolebinding riff-dev-edit --namespace=${NAMESPACE} --clusterrole=edit --serviceaccount=${NAMESPACE}:riff-dev
# run the utils using the service account as a pod
kubectl run riff-dev --namespace=${NAMESPACE} --image=projectriff/dev-utils --serviceaccount=riff-dev --generator=run-pod/v1
```

As pods do not survive node failures, over time the riff-dev pod may stop running. When this happens create a new pod using the same service account.

## Included tools
1. **publish:** To publish an event to the given stream.
The command takes the form:
    ```
    publish <stream-name> -n <namespace> --payload <payload-as-string> --content-type <content-type> --header "<header-name>: <header-value>"
    ```
    where `stream-name`, `payload` and `content-type` are mandatory and `header` can be used multiple times.
1. **subscribe:** To subscribe for events from the given stream.
The command takes the form:
    ```
    subscribe <stream-name> --from-beginning
    ```
    If the `--from-beginning` option is present, display all the events in the stream, otherwise only new events are displayed in the following json format:
    ```
    {"payload": "base64 encoded user payload","content-type": "the content type of the message","headers": {"user provided header": "while publishing"}}
    ```
1. [jq](https://stedolan.github.io/jq/): To process JSON.

1. [base64](http://manpages.ubuntu.com/manpages/bionic/man1/base64.1.html): Encode and decode base64 strings.

1. [curl](https://curl.haxx.se/): To make HTTP requests.

1. [kafkacat](https://github.com/edenhill/kafkacat): To interact with kafka

The namespace parameter is optional for all the commands. If not specified, the namespace of the `riff-dev` pod will be assumed.

## Examples
These tools can be invoked using kubectl exec. some examples follow:

```bash
kubectl exec riff-dev --namespace ${NAMESPACE} -it -- publish letters --content-type text/plain --payload foo
```

```bash
kubectl exec riff-dev --namespace ${NAMESPACE} -it -- subscribe letters --from-beginning
```

```bash
kubectl exec riff-dev --namespace ${NAMESPACE} -it -- curl http://hello.default.svc.cluster.local/ -H 'Content-Type: text/plain' -H 'Accept: text/plain' -d '<insert your name>'
```

```bash
kubectl exec riff-dev --namespace ${NAMESPACE} -it -- bash
```

```bash
kafkacat -C -b YOUR_BROKER -t TOPIC
```