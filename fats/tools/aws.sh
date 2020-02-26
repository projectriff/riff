#!/bin/bash

base_url="${2:-https://s3.amazonaws.com/aws-cli}"

curl -s "${base_url}/awscli-bundle.zip" -o "awscli-bundle.zip"
unzip -qq awscli-bundle.zip
sudo ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws
