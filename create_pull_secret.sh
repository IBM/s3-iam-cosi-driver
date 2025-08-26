#!/bin/bash

# Define the namespace and secret name
NAMESPACE="s3-iam-cosi-driver"
SECRET_NAME="image-pull-secret"

# Check if the namespace exists
if ! kubectl get namespace $NAMESPACE &>/dev/null; then
  echo "Namespace '$NAMESPACE' does not exist. Creating it..."
  kubectl apply -f resources/ns.yaml
  if [[ $? -eq 0 ]]; then
    echo "Namespace '$NAMESPACE' created successfully."
  else
    echo "Error creating namespace '$NAMESPACE'. Exiting."
    exit 1
  fi
else
  echo "Namespace '$NAMESPACE' already exists."
fi

# Prompt the user for the email address
read -p "Enter your email address (for registry): " EMAIL

# Prompt the user for the API key
read -sp "Enter your API key for icr.io: " API_KEY
echo

# Validate inputs
if [[ -z "$EMAIL" || -z "$API_KEY" ]]; then
  echo "Error: Email and API key must be provided."
  exit 1
fi

# Define the registry server
REGISTRY_SERVER="icr.io"

# Create the Kubernetes secret in the specified namespace
kubectl create secret docker-registry $SECRET_NAME \
  --namespace=$NAMESPACE \
  --docker-server=$REGISTRY_SERVER \
  --docker-username=iamapikey \
  --docker-password="$API_KEY" \
  --docker-email="$EMAIL"

# Confirm creation
if [[ $? -eq 0 ]]; then
  echo "Secret '$SECRET_NAME' created successfully in namespace '$NAMESPACE'!"
else
  echo "Error creating the secret in namespace '$NAMESPACE'."
  exit 1
fi