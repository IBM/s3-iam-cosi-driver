#!/bin/sh

# Path to the mounted COSI BucketInfo file
COSI_FILE="/data/cosi/BucketInfo"

# Extract credentials and endpoint from BucketInfo
ACCESS_KEY=$(jq -r '.spec.secretS3.accessKeyID' "$COSI_FILE")
SECRET_KEY=$(jq -r '.spec.secretS3.accessSecretKey' "$COSI_FILE")
ENDPOINT=$(jq -r '.spec.secretS3.endpoint' "$COSI_FILE")
REGION=$(jq -r '.spec.secretS3.region' "$COSI_FILE")

# Default region if not provided
if [ -z "$REGION" ]; then
  REGION="us-east-1"
fi

# Configure AWS CLI
aws configure set aws_access_key_id "$ACCESS_KEY"
aws configure set aws_secret_access_key "$SECRET_KEY"
aws configure set default.region "$REGION"

# Define a shell function for S3 listing
cat <<EOF > /usr/local/bin/kaws
#!/bin/sh
aws s3 --endpoint-url "$ENDPOINT" --no-verify-ssl "\$@"
EOF

# Make the function executable
chmod +x /usr/local/bin/kaws

# Test command (optional)
echo "AWS CLI configured with endpoint: $ENDPOINT"
aws s3 ls --endpoint-url "$ENDPOINT" --no-verify-ssl