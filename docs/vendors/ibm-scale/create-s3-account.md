# Creating a CES S3 Account

The procedure is already documented in official product documentation.  These
instructions are just a guide to help you through the process.

- [Main section](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=protocol-managing-s3-accounts-buckets)
- [Create an account](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=buckets-managing-s3-accounts)

Here's the short list of steps:

```bash
# Create an account
mms3 account create ACCOUNT_NAME --newBucketsPath PATH --gid 6666 --uid 6666
```

This will generate the accces and secret key for the account and displays those
at the STDOUT.

The PATH is a path under a filesystem (i.e. /gpfs/fs1/account1).  If directory
`account1` does not exist, it will be created automatically.

## Managing S3 accounts

- [Managing S3 accounts](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=buckets-managing-s3-using-mms3-command)

Here's the short list of steps:

Here's the short list of commands:

```bash
# List account and its access and secret keys
mms3 account list ACCOUNT_NAME

# List all accounts
mms3 account list

# Delete an account
mms3 account delete ACCOUNT_NAME

# update an account (see help for more options)
mms3 account update ACCOUNT_NAME --newBucketsPath PATH
mms3 account update ACCOUNT_NAME --resetKey
mms3 account update ACCOUNT_NAME --accessKey string
mms3 account update ACCOUNT_NAME --secretKey string
```

## Creating Buckets

That's the purpose of the `s3-iam-cosi-driver` :)

