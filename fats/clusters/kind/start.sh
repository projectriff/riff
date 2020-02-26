#!/bin/bash

cat <<EOF > ${CLUSTER_NAME}.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
EOF

if [ "$REGISTRY" = "docker-daemon" ] ; then
  registry_ip=$(docker inspect --format "{{.NetworkSettings.IPAddress }}" registry)
  # patch cluster config for registry location
  cat <<EOF >> ${CLUSTER_NAME}.yaml
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry.kube-system.svc.cluster.local:5000"]
    endpoint = ["http://${registry_ip}:5000"]
EOF
fi

kind create cluster --name ${CLUSTER_NAME} \
  --config ${CLUSTER_NAME}.yaml \
  --image kindest/node:v1.15.7 \
  --wait 5m

if [ "$REGISTRY" = "docker-daemon" ] ; then
  docker exec ${CLUSTER_NAME}-control-plane bash -c "echo \"${registry_ip} registry.kube-system.svc.cluster.local\" >> /etc/hosts"
  sudo su -c "echo \"${registry_ip} registry.kube-system.svc.cluster.local\" >> /etc/hosts"

  cat <<EOF | kubectl create -f -
---
kind: Service
apiVersion: v1
metadata:
  name: registry
  namespace: kube-system
spec:
  ports:
  - protocol: TCP
    port: 5000
    targetPort: 5000
---
kind: Endpoints
apiVersion: v1
metadata:
  name: registry
  namespace: kube-system
subsets:
  - addresses:
    - ip: ${registry_ip}
    ports:
      - port: 5000
EOF
fi

# move kubeconfig to expected location
mkdir -p ~/.kube
cp <(kind get kubeconfig --name ${CLUSTER_NAME}) ~/.kube/config
