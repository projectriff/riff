#!/bin/bash

eksctl create cluster --name $CLUSTER_NAME --version 1.10 --verbose 4
