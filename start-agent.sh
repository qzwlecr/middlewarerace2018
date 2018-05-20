#!/bin/bash

ETCD_HOST=etcd
ETCD_PORT=2379
ETCD_URL=http://$ETCD_HOST:$ETCD_PORT

echo ETCD_URL = $ETCD_URL

if [[ "$1" == "consumer" ]]; then
  echo "Starting consumer agent..."

elif [[ "$1" == "provider-small" ]]; then
  echo "Starting small provider agent..."

elif [[ "$1" == "provider-medium" ]]; then
  echo "Starting medium provider agent..."

elif [[ "$1" == "provider-large" ]]; then
  echo "Starting large provider agent..."

else
  echo "Unrecognized arguments, exit."
  exit 1
fi