# Multiple Tenants

The `s3-iam-cosi-driver` supports multi-tenancy where multiple tenants can share the same Object Storage Provider (OSP). Each tenant is mapped to a unique S3 account ensuring isolation between tenants.

For a new tenant account, the following workflow describes setting up multi-tenancy:

1. [OSP Administrator](./roles.md) creates S3 account for the tenant
   - Configures appropriate storage quotas and limits
   - Generates account access and secret keys
2. [OSP Administrator](./roles.md) provides credentials to [Tenant Administrator](./roles.md)
3. [Tenant Administrator](./roles.md) sets up COSI driver using the provided credentials:
   - Creates Secret with S3 Account credentials
   - Creates BucketClass with Secret reference
   - Creates BucketAccessClass with Secret reference
   - Note: this process is described in [COSI design](./introduction.md)

This architecture ensures:

- Tenant isolation through separate S3 accounts
- Independent IAM user management per tenant
- Separate bucket policy management
- Distinct credential management

Each tenant can manage their own buckets and access controls without affecting other tenants in the system.
