# Introduction to `s3-iam-cosi-driver`

The `s3-iam-cosi-driver` is a generic COSI implementation designed for Object Storage Providers (OSPs) that support:

- S3-compatible APIs
- IAM user management
- Bucket policy-based access control

This driver can work with various storage providers such as:

- RedHat Noobaa
- IBM Storage Scale
- Amazon S3
- Other S3-compatible storage systems with IAM-like authentication

For details about the COSI standard, see:

- [COSI Overview](https://kubernetes.io/blog/2022/09/02/cosi-kubernetes-object-storage-management/)
- [COSI Standard](https://github.com/kubernetes/enhancements/tree/master/keps/sig-storage/1979-object-storage-support)

## Architecture

The `s3-iam-cosi-driver` follows the standard COSI workflow.  It consists of an
[K8s Admin](./roles.md) setting up the BucketClass and BucketAccessClass CRs
(one time setup).  Then [Users](./roles.md) can provision buckets and request
bucket access through BucketClaim and BucketAccess CRs, respectively.

Workflow:

1. [Admin](./roles.md) creates BucketClass and BucketAccessClass
2. [User](./roles.md) creates BucketClaim
3. COSI Driver:
   - Watches for BucketClaim creation
   - Provisions bucket on Object Storage Provider
   - Creates Bucket object in Kubernetes API
4. [User](./roles.md) creates BucketAccess
5. COSI Driver:
   - Watches for BucketAccess creation
   - Manages (grant or deny) access permissions

For details on the workflow and YAMLs see:

- [Bucket Provisioning](./bucket-provisioning.md)
- [Bucket Access Management](./bucket-access-management.md)
