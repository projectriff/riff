#!/bin/bash

if [ "$REGISTRY" = "docker-daemon" ] ; then
  kubectl delete service registry -n kube-system
  kubectl delete endpoint registry -n kube-system
fi

kind delete cluster --name ${CLUSTER_NAME}
