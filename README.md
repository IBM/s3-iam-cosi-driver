[![Build Status](https://v3.travis.ibm.com/graphene/s3-iam-cosi-driver.svg?token=5BVjiEmGgixmiW4VYFEG&branch=main)](https://v3.travis.ibm.com/graphene/s3-iam-cosi-driver)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

# `s3-iam-cosi-driver`

## ⚠️ Disclaimer

This repository is intended for experimental and research use only.
It is not officially supported by IBM or any IBM product team.
Maintainers will provide support on a best-effort basis.

## Introduction

The `s3-iam-cosi-driver` is a generalized implementation of the COSI standard for
any S3 OSP (Object Storage Provider) that supports full IAM Users and Bucket Policies for
managing user access to buckets.  Since this driver is not vendor specific its designed
to worked with a variety of OSPs.

> For more information on the COSI standard, see:
https://kubernetes.io/blog/2022/09/02/cosi-kubernetes-object-storage-management/

## Prerequisites

### 1. Object Storage Provider (OSP)

The OSP must have:

- **S3 Endpoint**: Full URL of the S3 service (e.g., `https://s3.example.com` or `http://192.168.1.100`)
  - Must include the protocol (`http://` or `https://`)
  - Can be either a hostname or IP address
  - Should not include the port number (specified separately)
- **S3 Port**: Port number for the S3 service
- **IAM Port**: Port number for the IAM service
- **Account Name**: (optional) unique ID for the account in the OSP
- **Account Credentials**: Access & Secret keys for each S3 Account

> Note: S3 Account setup and credential generation may vary by vendor. See [vendor-specific instructions](./docs/vendors/) for details.

### 2. COSI Controller & CRDs

To install COSI CRDs and Controller, run the following:

```bash
mkdir container-object-storage-interface
git@github.com:kubernetes-sigs/container-object-storage-interface.git
cd container-object-storage-interface
kubectl apply -k .
```

Verify the COSI controller is running:

   ```bash
   kubectl get pods -n container-object-storage-system
   ```

   Expected output:

   ```bash
   NAME                                                   READY   STATUS    RESTARTS   AGE
   container-object-storage-controller-7f9f89fd45-pjhh6   1/1     Running   0          38m
   ```

> See official documentation https://github.com/kubernetes-sigs/container-object-storage-interface

## Building `s3-iam-cosi-driver` Images

The `s3-iam-cosi-driver` images are automatically built and pushed to `icr.io` through our CI/CD pipeline for each commit or merge into the `main` branch.  Manual building is typically not necessary.

If you need to build the images manually, please refer to our [build instructions](./BUILD.md).

## Setup Pull Secret

Run the following to setup the pull secret in ICR

Note: you must have API key from ICR

```bash
./create_pull_secret.sh
```

Here's what you can expect in the output:

```bash
❯ ./create_pull_secret.sh
Namespace 's3-iam-cosi-driver' does not exist. Creating it...
namespace/s3-iam-cosi-driver created
Namespace 's3-iam-cosi-driver' created successfully.
Enter your email address (for registry): myemail@acme.com
Enter your API key for icr.io:
```

## Installing `s3-iam-cosi-driver`

Now start the sidecar and cosi driver with:

```bash
kubectl create -k resources/
```

This should result in the COSI driver running:

```bash
❯ kubectl -n s3-iam-cosi-driver get pods
NAME                                        READY   STATUS    RESTARTS   AGE
s3-iam-cosi-provisioner-6d9dfcb77-jv9g4   2/2     Running   0          3m34s
```

## Setup Guide

### Administrator Setup

1. Create Account Secret

   ```bash
   # Create a secret for each S3 Account using the provided template
   cp examples/bucketsecret-template.yaml examples/bucketsecret.yaml
   # Edit examples/bucketsecret.yaml with your account details
   kubectl create -f examples/bucketsecret.yaml
   ```

2. Configure Storage Classes

   ```bash
   # Create the bucket and access classes
   kubectl create -f examples/bucketclass.yaml
   kubectl create -f examples/bucketaccessclass.yaml
   ```

   > Note: The BucketClass represents an S3 Account configuration.

### User Setup

Once the administrator has completed the setup, users can:

1. Create and Access Buckets

   ```bash
   # Create a new bucket
   kubectl create -f examples/bucketclaim.yaml

   # Request access to the bucket
   kubectl create -f examples/bucketaccess.yaml
   ```

2. Verify Setup (Optional)

   ```bash
   # Deploy a test pod with AWS CLI
   kubectl create -f examples/awscliapppod.yaml

   # Verify bucket credentials are mounted correctly
   kubectl exec -it awscli -- cat /data/cosi/BucketInfo
   ```

   Example output:

   ```json
   {
     "metadata": {
       "name": "bc-b31eab82-85ba-40bc-8908-e937f2684015",
       "creationTimestamp": null
     },
     "spec": {
       "bucketName": "account1-bc189245ac-5ebd-4fa5-b24e-2e0dafeff478",
       "authenticationType": "KEY",
       "secretS3": {
         "endpoint": "http://9.46.74.154:6001",
         "region": "",
         "accessKeyID": "EXAMPLE_SECRET_KEY",
         "accessSecretKey": "EXAMPLE_SECRET_KEY"
       },
       "secretAzure": null,
       "protocols": ["s3"]
     }
   }
   ```

> Note: All buckets created under a BucketClass will be created within the same S3 account.

## Support

For issues and feature requests:
- File an issue in our [issue tracker]()
- Join our [community channel]()
- Check our [documentation]()

## License

This project is derived from the Ceph-COSI driver and contains code under two licenses:

1. Original code from Ceph-COSI driver: [Apache License 2.0](./LICENSE.apache2)
2. New modifications and additions: [MIT License](./LICENSE)

Both licenses are permissive open source licenses that allow for commercial use, modification, and distribution. The Apache 2.0 license includes explicit patent grants and trademark usage terms, while the MIT license is simpler but provides similar freedoms.

For detailed terms, please see the respective license files:

- [Apache License 2.0](./LICENSE.apache2)
- [MIT License](./LICENSE)