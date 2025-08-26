# Grant Access to Buckets

After the bucket is [created](./create-buckets.md), the `s3-iam-cosi-driver` ensures that all IAM users are denied access to the bucket by default.  In the COSI standard workflow, bucket access is handled through BucketAccess CRs.

When a BucketAccess CR is created, the COSI driver will verify against the BucketAccessClass to ensure the access request is allowed.  If the access request is allowed, the COSI driver will create an IAM user and attach the requested bucket policy to the bucket.  From the generated IAM user credentials, it will then store the access & secret keys for the IAM user in the Secret specified by `credentialsSecretName`.

In order to faciliate the request, the COSI driver requires the following fields from the BucketAccess CR:

- `bucketClaimName` - from this we derive the bucket name that users is requesting access
- `accessRequest` - the permissions being requested by the user
- `credentialsSecretName` - where the user wants the credentials stored (if granted)

Additionally, COSI driver requires the following from the BucketAccessClass:

- `iamUserPattern` - used to derive the IAM user name.  This must be unique within the S3 Account.

If access is denied, the COSI driver will not create an IAM user and the BucketAccess CR will be marked `AccessGranted:false`.

If access is granted, the COSI driver proceeds with the following steps:

```bash
# Create IAM user associated with BucketAccess CR
❯ aws-account1-iam create-user --user-name Frank
{
    "User": {
        "Path": "/",
        "UserName": "Frank",
        "UserId": "67bf803b01df245cb15f5ca7",
        "Arn": "arn:aws:iam::679bd45a926dc76c8dbcec26:user/Frank",
        "CreateDate": "2025-02-26T20:57:31+00:00"
    }
}

# Generate the access & secret access keys for the IAM user
# (these credentials will be provided in a Secret specified
# by the credentialsSecretName field in the BucketAccess CR)
❯ aws-account1-iam create-access-key --user-name Frank
{
    "AccessKey": {
        "UserName": "Frank",
        "AccessKeyId": "Oet1px30phy7q2XIOtph",
        "Status": "Active",
        "SecretAccessKey": "EXAMPLE_SECRET_KEY",
        "CreateDate": "2025-02-26T20:57:51+00:00"
    }
}

# Note, Frank cannot access the bucket yet
❯ aws-frank s3 ls
❯ echo $?
0

# Verify IAM user doesn't already have access
❯ aws-account1 s3api get-bucket-policy --bucket bucket1 | jq
{
  "Policy": {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Deny",
        "Principal": "*",
        "Action": "s3:*",
        "Resource": [
          "arn:aws:s3:::bucket1",
          "arn:aws:s3:::bucket1/*"
        ]
      }
    ]
  }
}

# Get UserID, this is needed for the policy (see [add-access.json](./add-access.json))
❯ aws-account1-iam list-users
{
    "Users": [
        {
            "Path": "/",
            "UserName": "Frank",
            "UserId": "682e4c8ac7d31535b9f53582",
            "Arn": "arn:aws:iam::679bd45a926dc76c8dbcec26:user/Frank",
            "CreateDate": "2025-05-21T21:58:34+00:00",
            "PasswordLastUsed": "2025-05-21T21:58:43+00:00"
        }
    ]
}

# Add IAM user to bucket policy with requested permissions
# (here let's assume BucketAccess CR has requested "s3:*" permissions
# and this was allowed by admin's BucketAccessClass)
❯ aws-account1 s3api put-bucket-policy --bucket bucket1 --policy file://add-access.json

# Verify the bucket policy
❯ aws-account1 s3api get-bucket-policy --bucket bucket1 | jq
{
  "Policy": {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "AWS": ["User-ID-for-Frank"]
        },
        "Action": "s3:*",
        "Resource": ["arn:aws:s3:::bucket1", "arn:aws:s3:::bucket1/*"]
      }
    ]
  }
}

# Verify IAM user has access
❯ aws-frank s3 ls
2025-02-26 13:21:43 bucket1

# Verify other IAM users do not have access
❯ aws-bob s3 ls
2025-02-03 12:30:24 bill-bucket
2025-02-03 12:30:27 bob-bucket
2025-02-03 12:30:17 shared-bucket
```
