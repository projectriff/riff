#!/bin/bash

helm reset
kubectl delete serviceaccount tiller -n kube-system
kubectl delete clusterrolebinding tiller
