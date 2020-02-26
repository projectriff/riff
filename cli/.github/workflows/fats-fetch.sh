#!/bin/bash

dir=${1}
refspec=${2:-master}
repo=${3:-projectriff/fats}

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/${repo}/archive/${refspec}.tar.gz > fats.tar.gz
  tar -xzf fats.tar.gz -C $dir --strip-components 1
  rm fats.tar.gz
fi
