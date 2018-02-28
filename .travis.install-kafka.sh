#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

sudo apt-get -qq update
sudo apt-get install -y default-jre zookeeperd
mkdir -p ~/Downloads
wget "https://archive.apache.org/dist/kafka/0.11.0.1/kafka_2.11-0.11.0.1.tgz" -O ~/Downloads/kafka.tgz
mkdir -p ~/kafka
( cd ~/kafka ; tar -xvzf ~/Downloads/kafka.tgz --strip 1 )
# echo "delete.topic.enable = true" >> ~/kafka/config/server.properties
nohup ~/kafka/bin/kafka-server-start.sh ~/kafka/config/server.properties > ~/kafka/kafka.log 2>&1 &
sleep 5
