#!/bin/bash

echo "RESOURCES:"
kubectl get deployments,services,pods,cm,sa,secrets,riff --all-namespaces || true
echo "RIFF RESOURCES:"
kubectl describe riff --all-namespaces || true
echo "NON-RUNNING PODS:"
kubectl get pods --all-namespaces --field-selector=status.phase!=Running \
| tail -n +2 | awk '{print "-n", $1, $2}' | xargs -L 1 kubectl describe pod || true
echo "NODES:"
kubectl describe node || true
