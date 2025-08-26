#!/bin/bash

aws_account1="PAGER=cat AWS_ENDPOINT_URL=https://9.46.85.238:6443 aws --profile nc-account1 --ca-bundle=~/.aws/certs/jc-c1/tls.crt"
aws_account1_iam="PAGER=cat AWS_ENDPOINT_URL=https://9.46.85.238:7005 aws iam --profile nc-account1 --ca-bundle=~/.aws/certs/jc-c1/tls.crt"

bucket_name="account1-bc66c23938-ac36-4854-8600-4ffce41985d0"

echo "--- Bucket Policy ---"
eval $aws_account1 s3api get-bucket-policy --bucket $bucket_name | jq -r .Policy | jq
echo -e "\n"

echo "--- IAM Users ---"
eval $aws_account1_iam list-users
echo -e "\n"

# Loop through all users and list their access keys
for user in $(eval $aws_account1_iam list-users | jq -r '.Users[].UserName'); do
  echo "--- IAM User: $user ---"
  eval $aws_account1_iam list-access-keys --user $user
  echo -e "\n"
done
