#!/bin/bash
set -e

NAME="app"
ADD="10"

data="$(kubectl get configmap "$NAME" -o json)"
max="$(echo "$data" | jq -rc '.data.max')"
size="$(echo "$data" | jq -rc '.data.size')"

size="$((size + ADD))"
if [[ "$size" -gt "$max" ]];then
  size="$max"
fi

pods=($(kubectl get pods -l app="$NAME" -o name))
count="${#pods[@]}"
pod_size="$((size/count))"

for pod in "${pods[@]}";do
  echo "$pod - $pod_size"
  kubectl exec "$pod" -- /bin/sh -c \
    "dd if=/dev/zero of=/tmp/data.tmp bs=1M count=$pod_size;mv /tmp/data.tmp /tmp/data"
done

kubectl patch configmap "$NAME" -p "{\"data\":{\"size\":\"$size\"}}"
