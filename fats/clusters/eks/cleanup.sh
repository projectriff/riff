#!/bin/bash

# delete job resources

eksctl delete cluster --name $CLUSTER_NAME --verbose 4
