#!/bin/bash

dir=${1}
refspec=${2:-3c0fa2533d4fcbc93cba485274562f4c37f3d015} # projectriff/fats fats2 as of 2018-12-08

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
