#!/bin/bash

riff_version="${1:-latest}"
base_url="${2:-https://storage.googleapis.com/projectriff/riff-cli/releases}"

if [ "$machine" == "MinGw" ]; then
  curl -L ${base_url}/${riff_version}/riff-windows-amd64.zip > riff.zip
  unzip riff.zip -d /usr/bin/
  rm riff.zip
else
  riff_dir=`mktemp -d riff.XXXX`

  curl -L ${base_url}/${riff_version}/riff-linux-amd64.tgz | tar xz -C $riff_dir
  chmod +x $riff_dir/riff
  sudo mv $riff_dir/riff /usr/local/bin/

  rm -rf $riff_dir
fi
