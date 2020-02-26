#!/bin/bash

url=$1
curl_opts=$2
expected_data=$3

echo "Invoke $url"

ip=$(kubectl get service -n projectcontour envoy-external -o jsonpath='{$.status.loadBalancer.ingress[0].ip}')
port="80"
if [ -z "$ip" ]; then
  ip=$(kubectl get node -o jsonpath='{$.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
  if [ -z "$ip" ] ; then
    ip=$(kubectl get node -o jsonpath='{$.items[0].status.addresses[?(@.type=="InternalIP")].address}')
  fi
  if [ -z "$ip" ] ; then
    ip=localhost
  fi
  port=$(kubectl get service -n projectcontour envoy-external -o jsonpath='{$.spec.ports[?(@.name=="http")].nodePort}')
fi

hostname=$(echo "$url" | sed -e 's|http://||g')
curl "http://${ip}:${port}/" -H "Host: ${hostname}" $curl_opts -v | tee curl.out

actual_data=$(cat curl.out | tail -1)
rm curl.out

# add a new line after invoke, but without impacting the curl output
echo ""

fats_assert "$expected_data" "$actual_data"
