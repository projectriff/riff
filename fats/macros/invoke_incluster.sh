#!/bin/bash

url=$1
curl_opts=$2
expected_data=$3

echo "Invoke $url"

kubectl exec riff-dev -n $NAMESPACE -- curl $url $curl_opts -v | tee curl.out

actual_data=$(cat curl.out | tail -1)
rm curl.out

# add a new line after invoke, but without impacting the curl output
echo ""

fats_assert "$expected_data" "$actual_data"
