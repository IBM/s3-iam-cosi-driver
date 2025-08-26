# Roles

The following roles are defined for the `s3-iam-cosi-driver`:

## OSP Administrator

The OSP (Object Storage Provider) Administrator has administrative access to the Object Storage service or system. They are responsible for managing and maintaining S3 accounts.

## Kubernetes Administrator

The Kubernetes Administrator has administrative access to the Kubernetes cluster. In the context of COSI drivers, they have the ability to create BucketClass and BucketAccessClass CRs.

## Appplication Developer / Data Scientist / End User

This user has namespace-scoped access to the Kubernetes cluster. In the context of COSI drivers, they have the ability to create BucketClaim and BucketAccess CRs within their respective namespaces.

## COSI Driver

The COSI Driver is the component that watches for the creation of BucketClaim and BucketAccess CRs and takes the appropriate action to provision buckets and grant access to users.

## Site Admin

Used in the content of multi-tenancy.  This is the highest-level administrator who:

- Has overall responsibility for the entire system
- Manages both the client infrastructure and storage infrastructure
- Has the authority to delegate control to Tenant Admins
- Acts as the root-level administrator for multi-tenant environments

Think of the Site Admin as the "landlord" of the entire system.

## Tenant Admin

This is a delegated administrator who:

- Manages resources for a specific tenant/group
- Receives their access and control permissions from the Site Admin
- Can only manage resources that belong to their specific tenant
- Has limited scope compared to the Site Admin

Think of the Tenant Admin as an "apartment manager" who only has authority over their specific section of the building.
