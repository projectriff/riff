#!/bin/bash

echo "Remove riff and Knative resources"
kubectl delete riff --all-namespaces --all
kubectl delete knative --all-namespaces --all
