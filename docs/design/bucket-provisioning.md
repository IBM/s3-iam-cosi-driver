# Bucket Provisioning

The worfklow follows a similar pattern as StorageClass and PersistentVolumeClaim for dynamic provisioning.  Here is the workflow for the `s3-iam-cosi-driver`:

1. The [K8s Admin](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md) creates the Account Secret.
2. The [K8s Admin](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md) creates the BucketClass.
3. The [User](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md) creates a BucketClaim.

After creating the BucketClaim, the `s3-iam-cosi-driver` will:

1. Create a Bucket CR - which represents the bucket resource and status of the bucket on the Object Storage Provider
2. Issue the command to create the actual bucket on the Object Storage Provider

The following sections describe the CRDs and the workflow in detail.

## Account Secret

The OSP details are stored in a Secret and is created by the [Kubernetes Admin](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md) for each S3 account.

Here's the expect format for the Secret YAML:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: s3-account1
  namespace: s3-iam-cosi-driver
type: Opaque
data:
  # all values shoule be encoded in base64 (not shown here)
  Endpoint: https://s3.example.com
  S3Port: 6443
  IAMPort: 7004
  AccountName: s3-account-1  # optional
  AccessKey: abc123
  SecretKey: abc123
```

## BucketClass

The BucketClass CR is created by the [K8s Admin](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md).

Here's the expect format for the BucketClass YAML:

```yaml
apiVersion: objectstorage.k8s.io/v1alpha1
kind: BucketClass
metadata:
  name: s3-account1-bc
driverName: s3-iam.objectstorage.k8s.io
deletionPolicy: [Delete|Retain]
parameters:
  # S3 Account Secret & Namespace
  # +required
  # +s3-iam-cosi
  accountSecret: s3-account-1
  accountSecretNamespace: s3-iam-cosi-driver
```

## BucketClaim

The BucketClaim CR is created by the [User](https://github.ibm.com/graphene/s3-iam-cosi-driver/blob/main/docs/design/roles.md).

The BucketClaim will request a bucket from the COSI Driver.  The COSI Driver
will then provision the bucket on the Object Storage provider.

Here's the expect format for the BucketClaim YAML:

```yaml
kind: BucketClaim
apiVersion: objectstorage.k8s.io/v1alpha1
metadata:
  name: my-bucket1
  namespace: default
spec:
  # Name of the BucketClass
  # +required
  # +cosi:standard
  bucketClassName: s3-account1-bc

  # The protocols supported by the bucket, only supports S3 for now.
  # COSI standard allows for other protocols such as Azure Blob Storage and Google Cloud Storage.
  # +required
  # +cosi:standard
  protocols:
  - s3
```

## Bucket

The Bucket object is created by the COSI Driver when the BucketClaim is created.

Here's an example of the Bucket object:

```bash
❯ kubectl describe bucket account1-bc84b873f7-658a-4304-8bd0-e8aafc731bee
Name:         account1-bc84b873f7-658a-4304-8bd0-e8aafc731bee
Namespace:
Labels:       <none>
Annotations:  <none>
API Version:  objectstorage.k8s.io/v1alpha1
Kind:         Bucket
Metadata:
  Creation Timestamp:  2025-01-15T22:12:58Z
  Finalizers:
    cosi.objectstorage.k8s.io/bucket-protection
    cosi.objectstorage.k8s.io/bucketaccess-bucket-protection
  Generation:        1
  Resource Version:  3567515
  UID:               695f94d9-1395-49ea-9d08-809b567c9cc8
Spec:
  Bucket Claim:
    Name:             my-bucket1
    Namespace:        default
    UID:              84b873f7-658a-4304-8bd0-e8aafc731bee
  Bucket Class Name:  account1-bc
  Deletion Policy:    Delete
  Driver Name:        cosi.s3.iam.objectstorage.k8s.io
  Parameters:
    Account Secret:            s3-account1
    Account Secret Namespace:  s3-iam-cosi-driver
  Protocols:
    s3
Status:
  Bucket ID:     account1-bc84b873f7-658a-4304-8bd0-e8aafc731bee
  Bucket Ready:  true
Events:          <none>
```

The Bucket object contains the following fields:

- Bucket ID: The ID of the bucket.  This is the same name of the actual bucket on the Object Storage provider.
- Bucket Ready: Whether the bucket is created and ready to be accessed by the User.
