#!/bin/bash

set -o errexit
set -o pipefail

riff_version=`cat VERSION`
helm_charts_bucket='riff-charts'
helm_charts_url="https://${helm_charts_bucket}.storage.googleapis.com/"
work_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/helm-charts/.work"

if [[ "$GCLOUD_CLIENT_SECRET" != "" ]]; then
  echo $GCLOUD_CLIENT_SECRET | base64 --decode > client-secret.json
  gcloud auth activate-service-account --key-file client-secret.json
  rm client-secret.json
fi

pushd helm-charts
  helm init --client-only

  mkdir -p $work_dir
  gsutil cp "gs://$helm_charts_bucket/index.yaml" $work_dir/

  if [[ $riff_version != *-snapshot ]]; then
    echo "Setting latest version to $riff_version"
    echo $riff_version > $work_dir/latest_version
  fi

  sed -E -i '' "s/tag:[ ]*latest/tag: $riff_version/" riff/values.yaml
  sed -E -i '' "s/\|[ ]*latest[ ]*\|/|$riff_version|/" riff/README.md
  helm package riff --version "$riff_version" --app-version "$riff_version" --destination $work_dir
  helm repo index $work_dir --url "$helm_charts_url" --merge $work_dir/index.yaml

  gsutil cp -a public-read "$work_dir/*.tgz" "gs://$helm_charts_bucket"
  gsutil cp -a public-read "$work_dir/index.yaml" "gs://$helm_charts_bucket"
  if [[ -f "$work_dir/latest_version" ]]; then
    gsutil cp -a public-read "$work_dir/latest_version" "gs://$helm_charts_bucket"
  fi
popd
