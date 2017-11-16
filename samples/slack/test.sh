#!/bin/bash

set -exo pipefail

clean() {
  helm --tiller-namespace sk8s del --purge sk8s || true
  (kubectl delete ns sk8s --cascade=true && while [ ! -z "$(kubectl get ns | grep sk8s)" ]; do sleep 10; done) || true
}

clean
if [ "$1" == "clean" ]; then
  exit 0
fi
kubectl create ns sk8s
helm init --tiller-namespace sk8s && sleep 10
helm repo update
helm --tiller-namespace sk8s install sk8srepo/sk8s --namespace sk8s -n sk8s
docker build -t sk8s/slack-command:test slack-command
docker push sk8s/slack-command:test
kubectl -n sk8s apply -f slack-command/slack-command.yaml
while [ -z "$gw_ip" ]; do
  sleep 10
  gw_ip=$(kubectl -n sk8s get svc -l component=http-gateway -o jsonpath='{.items[0].status.loadBalancer.ingress[].ip}')
done
gw_port=$(kubectl -n sk8s get svc -l component=http-gateway -o jsonpath='{.items[0].spec.ports[?(@.name == "http")].port}')
curl -d 'token=WFftWLOQvxVZa51jE7XHPku4&team_id=T024LQKAS&team_domain=pivotal&channel_id=D278SFTKN&channel_name=directmessage&user_id=U278V2UCE&user_name=cklassen&command=%2Fsk8s-gke&text=test&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT024LQKAS%2F273602624114%2FlL8zoKUxr5KnG7g4mBr9L7eG' http://$gw_ip:$gw_port/requests/slack
