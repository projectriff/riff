#!/bin/bash

utils_version=0.5.0-snapshot-20200203204409-43600a71d7f55c67

kubectl create serviceaccount riff-dev --namespace=${NAMESPACE}
kubectl create rolebinding riff-dev-edit --namespace=${NAMESPACE} --clusterrole=edit --serviceaccount=${NAMESPACE}:riff-dev
kubectl run riff-dev --namespace=${NAMESPACE} --image=projectriff/dev-utils:${utils_version} --serviceaccount=riff-dev --generator=run-pod/v1
kubectl wait pods riff-dev --for=condition=Ready --namespace=$NAMESPACE --timeout=60s
