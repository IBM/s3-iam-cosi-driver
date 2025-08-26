#!/bin/bash

# Array of COSI resources
resources=(
  "bucketclasses.objectstorage.k8s.io"
  "bucketaccessclasses.objectstorage.k8s.io"
  "bucketclaims.objectstorage.k8s.io"
  "bucketaccesses.objectstorage.k8s.io"
  #"objectbucketclaims.objectbucket.io"
  #"objectbuckets.objectbucket.io"
  "buckets.objectstorage.k8s.io"
)

echo "Fetching all COSI resources..."
echo "--------------------------------"

# Loop through each resource and run `kubectl get`
for resource in "${resources[@]}"; do
  echo "$resource"
  kubectl get "$resource" -A || echo "Failed to fetch $resource or resource does not exist."
  echo
done

echo "Test Pods"
echo "--------------------------------"
kubectl get pods -n default awscli
