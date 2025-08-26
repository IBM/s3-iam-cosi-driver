# Bucket Access Management

The COSI Driver will watch for the creation of BucketAccess CRs and take the appropriate
action to grant or deny access to the user.

Here is the overall workflow:

1. The [Administrator](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md#administrator) creates the BucketAccessClass CR
2. The [Data Scientist / User](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md#data-scientist-user) creates a BucketAccess CR

After creating the BucketAccess CR the COSI driver will grant or deny access to the user.

The following sections describe the CRDs and the workflow in detail.

## BucketAccessClass

The BucketAccessClass CR is created by the [Administrator](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md#administrator).

```yaml
kind: BucketAccessClass
apiVersion: objectstorage.k8s.io/v1alpha1
metadata:
  name: account1-bac
driverName: s3-iam.objectstorage.k8s.io

# Authentication Type
# Note: only supports KEY authentication.  See [Issue #9](https://github.ibm.com/graphene/s3-iam-cosi-driver/issues/9) for support for IAM.
# +required
# +cosi:standard
authenticationType: KEY

parameters:
  # S3 Account Secret
  # See [Account Secret](./bucket-provisioning.md#account-secret)
  # +required
  # +s3-iam-cosi
  accountSecret: s3-account-1
  accountSecretNamespace: s3-iam-cosi-driver

  # Unique IAM user name pattern
  # This pattern allows admins to construct IAM user names dynamically
  # Supported placeholders:
  #   ${bucketAccessName} - The name of the BucketAccessRequest
  #   ${namespace} - The namespace of the BucketAccessRequest
  #   ${random} - A random string for uniqueness
  # +required
  # +s3-iam-cosi
  iamUserPattern: "cosi-${namespace}-${bucketAccessName}"

  # The following sections are policies for granting or denying access to the bucket.
  # By default, all actions are denied unless they are explicitly allowed through
  # the policies defined below.  The user requests access through BucketAccess CRs.
  #
  # If user does NOT set the accessRequest field, then the defaultPolicy is
  # automatically granted.
  #
  # If the user sets the accessRequest field in the BucketAccess CR,
  # then the requested access is tested against the requestPolicy in this CR.
  # A request will be:
  # - Denied if *any* requested action appears in requestPolicy.deny
  # - Allowed if *all* requested actions appear in requestPolicy.allow
  # - Otherwise the request is denied (following the principle of all or nothing)
  #

  # The default policy when requesting access.
  # This is used only if the user doesn't specify the accessRequest in the
  # BucketAccess CR.  If Admin doesn't set any default policies, then the
  # default is to deny all actions (following the principle of security first).
  # +s3-iam-cosi
  defaultPolicy:

    # S3 actions allowed by default.
    # This must not conflicts with defaultPolicy.deny
    # "s3:*" here just symbolizes one or more S3 actions.
    # +optional
    allow:
       actions:
         - "s3:*"

    # S3 actions denied by default.
    # This must not conflicts with defaultPolicy.allow
    # If ommitted, any actions not explicitly allowed will automatically get
    # denied so this field mostly exists for compliance.
    # "s3:*" here just symbolizes one or more S3 actions.
    # +optional
    deny:
       actions:
         - "s3:*"

  # The policy when requesting access.
  # This is used only if the user sets the accessRequest field in the
  # bucketAccessCR.  If not, then defaultPolicy is used.
  # +s3-iam-cosi
  requestPolicy:
    # S3 actions allowed if requested.
    # This must not conflict with requestPolicy.deny
    # "s3:*" here just symbolizes one or more S3 actions.
    # +optional
    allow:
       actions:
        - "s3:*"

       # Expirations on allowed actions
       # +optional
       # +s3-iam-cosi
       timeToLive: "1h"

    # S3 actions denied if requested.
    # This also must not conflict with requestPolicy.allow.
    # If ommitted, any actions not explicitly allowed will automatically get
    # denied so this field mostly exists for compliance.
    # +optional
    deny:
       actions:
        - "s3:*"

  # Expirations on the entire bucket access
  # +optional
  # +s3-iam-cosi
  timeToLive: "2h"  # autorevoke after 2 hours


```

## BucketAccess

The BucketAccess CR is created by the [Data Scientist / User](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md#data-scientist-user).

```yaml
kind: BucketAccess
apiVersion: objectstorage.k8s.io/v1alpha1
metadata:
  name: my-bucket1-access
  namespace: default
spec:
  # BucketClaimName is the name of the BucketClaim.
  # +required
  # +s3-iam-cosi
  bucketClaimName: my-bucket1

  # Name of the BucketAccessClass
  # +required
  # +s3-iam-cosi
  bucketAccessClassName: account1-bac

  # Name of the Secret containing the bucket access credentials.
  # If access is granted, this Secret will contain the credentials.
  # +required
  # +s3-iam-cosi
  credentialsSecretName: my-bucket1-credentials

  # The protocols supported by the bucket, only supports S3 for now.
  # COSI standard allows for other protocols such as Azure Blob Storage and Google Cloud Storage
  # +optional
  # +s3-iam-cosi
  protocol: s3

  parameters:
    # The access being requested
    # If not specified, the default policy will be used from the BucketAccessClass.
    # +optional
    # +s3-iam-cosi
    accessRequest:
      # The actions being requested by the user.
      # "s3:*" here just symbolizes one or more S3 actions.
      # +optional
      # +s3-iam-cosi
      actions:
        - "s3:*"

      # The expiration time of the access request.
      # +optional
      # +s3-iam-cosi
      timeToLive: "2h"

```
