#!/bin/bash

dir=${1}
refspec=${2:-b78f390912581fd6d2b7ae62e0e9774bec4b18d2} # projectriff/fats v0.2.x as of 2019-02-06

if [ ! -f $dir ]; then
  mkdir -p $dir
  curl -L https://github.com/projectriff/fats/archive/${refspec}.tar.gz | \
    tar xz -C $dir --strip-components 1
fi
