#!/bin/bash

# Allow for insecure registries as long as docker daemon is actually running
echo '{ "insecure-registries": [ "registry.kube-system.svc.cluster.local:5000" ] }' | sudo tee /etc/docker/daemon.json > /dev/null
sudo systemctl daemon-reload
sudo systemctl restart docker

# Enable local registry
echo "Installing a daemon registry"
docker run -d -p 5000:5000 --name registry registry:2 || docker start registry
