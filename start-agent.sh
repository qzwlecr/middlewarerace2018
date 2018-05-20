#!/bin/bash

ETCD_HOST=etcd
ETCD_PORT=2379
ETCD_URL=http://$ETCD_HOST:$ETCD_PORT

echo ETCD_URL = $ETCD_URL

if [[ "$1" == "consumer" ]]; then
  echo "Starting consumer agent..."
  /root/dists/agent -m=1536 -t=consumer

elif [[ "$1" == "provider-small" ]]; then
  echo "Starting small provider agent..."
  /root/dists/agent -m=768 -t=provider -n=small

elif [[ "$1" == "provider-medium" ]]; then
  echo "Starting medium provider agent..."
  /root/dists/agent -m=768 -t=provider -n=medium

elif [[ "$1" == "provider-large" ]]; then
  echo "Starting large provider agent..."
  /root/dists/agent -m=768 -t=provider -n=large

else
  echo "Unrecognized arguments, exit."
  exit 1
fi
