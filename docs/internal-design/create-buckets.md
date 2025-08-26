# Creating Buckets with No-Access Policy

In the COSI standard workflow, buckets are created through BucketClaim CRs.
This triggers the `s3-iam-cosi-driver` to create the bucket. Following security
best practices, when creating buckets in the OSP, the driver must immediately
attach a policy to deny access to all users. This ensures bucket access is
controlled exclusively through BucketAccess CRs.

This is the procedure that the `s3-iam-cosi-driver` follows to create a bucket.  This
shows the manual steps that are performed, except the the driver will perform these
steps in code.

In this example, `account1` is the Root Account and `bucket1` is the bucket that will be created.  The policy being attached here can be found in [no-access.json](./no-access.json), which denies access to all users on the newly created bucket.  At that point, bucket access will be handled *only* through BucketAccess CRs.

```bash
# Create the bucket
❯ aws-account1 s3 mb s3://bucket1
make_bucket: bucket1

# Immediately attach the policy
❯ aws-account1 s3api put-bucket-policy --bucket bucket1 --policy file://no-access.json

# Verify the bucket policy
❯ aws-account1 s3api get-bucket-policy --bucket bucket1 | JQ
{
  "Policy": {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Deny",
        "Principal": "*",
        "Action": "s3:*",
        "Resource": ["arn:aws:s3:::bucket1", "arn:aws:s3:::bucket1/*"]
      }
    ]
  }
}
```

Here we can verify that other IAM users cannot access the bucket.

```bash
# List all buckets
❯ aws-account1 s3 ls
2025-02-26 13:21:43 bucket1  # <-- this is the bucket we created
2025-02-03 12:30:24 bill-bucket
2025-02-03 12:30:27 bob-bucket
2025-02-03 12:30:17 shared-bucket

# Try from account1 (root account)
# (root account should be able to access all buckets)
❯ aws-account1 s3 ls s3://bucket1
❯ echo $?
0

# Try from IAM user `bob`
# (bob should not be able to access bucket1)
❯ aws-bob s3 ls s3://bucket1
An error occurred (AccessDenied) when calling the ListObjectsV2 operation: Access Denied
```
