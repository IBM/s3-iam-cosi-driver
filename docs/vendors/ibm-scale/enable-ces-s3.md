# Enabling Scale CES S3

The procedure is already documented in official product documentation.  These
instructions are just a guide to help you through the process.

## Planning

- [System Requirements](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=s3-configuration-considerations-ces-clusters)
- [Each NooBaa endpoint forks a process](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=s3-configuration-considerations-each-noobaa-endpoint-fork-process)

## Installation & Upgrade

- [Installing Protocol Nodes](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=installing-storage-scale-linux-nodes-deploying-protocols)

- [Upgrading Protocol Nodes](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=upgrading-storage-scale-protocol-nodes)

## Configuring CES and CES IP addresses

- [Main section](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=configuring-ces-protocols)

- [Configuring Shared Root Filesystem](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=cces-setting-up-cluster-export-services-shared-root-file-system)

- [Configuring CES nodes](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=services-configuring-cluster-export-nodes)

- [Configuring IP Addresses](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=services-configuring-ces-protocol-service-ip-addresses)

- [Alias IP addresses](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=services-ces-ip-aliasing-network-adapters-protocol-nodes)

- [TLS Certificates](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=ccp-setting-up-self-signed-ssltls-certificates-secure-communication-between-s3-client-s3-service)

Here's the short list of steps:

```bash
# enable CES on the node
# (do this for each protocol node)
mmchnode -N NODE --ces-enable

# list CES nodes
mmces node list

# Add the alias IP address to the node
# Once added, CES will automatically assign the IP address to a node.
# To use IP aliasing, must run the following on all nodes
#    echo "IP_ADDRESS ces-ip-alias" | sudo tee -a /etc/hosts
mmces address add --ces-ip IP_ADDRESS

# List the CES addresses
mmces address list

# Optional: adjust address policy (see help)
mmces address policy even-coverage

# Verify final CES configuration
mmlscluster --ces
```

## Enabling S3 Service (Noobaa)

The `s3-iam-cosi-driver` requires IAM enabled in Noobaa.  This is not yet supported
by Scale CES using the mms3 command, so we need to do this manually.

To do this you'll need to modify the Noobaa config file in the shared root filesystem.
For example, if the shared root file system is mounted at `/ibm/fs0`, then
you would add this to the config.json file as follows:

```bash
# Edit the config.json file
vi /ibm/fs0/ces/s3-config/config.json

# Add the following to the config.json file
"ENDPOINT_SSL_IAM_PORT": 7005,
```

Note: config.json might not exist initially.  So you may have to start S3 service,
modify the config.json file, and then restart the S3 service.

Then continue to enable the S3 service as documented in the product documentation.

- [Enabling S3](https://www.ibm.com/docs/en/storage-scale/5.2.2?topic=services-configuring-enabling-smb-nfs-s3-protocol)

Here's the short list of steps:

```bash
# Enable S3 service
mmces service enable S3

# List the services
mmces service list

# Useful command to stop/start S3 on all nodes
mmces service [stop|start] S3 -a
```

Use the following command to control S3 ports used:

```bash
# List S3 configuration
mms3 config list

# List the S3 ports (officially supported by CES S3)
mms3  config list | grep PORT
 ENDPOINT_PORT : 6001
 ENDPOINT_SSL_PORT : 6443
```

Now check the ports that Noobaa is listening:

```bash
# sudo ss -tulpn | grep -i noobaa
tcp   LISTEN 0      511                *:7005             *:*    users:(("noobaa",pid=3255078,fd=25))
tcp   LISTEN 0      511                *:6443             *:*    users:(("noobaa",pid=3255078,fd=24))
tcp   LISTEN 0      511                *:6001             *:*    users:(("noobaa",pid=3255078,fd=23))
tcp   LISTEN 0      511                *:7004             *:*    users:(("noobaa",pid=3255078,fd=22))
```

In this example, the ports are:

- 6001 - S3 (http)
- 6443 - S3 (https)
- 7004 - S3 IAM
- 7005 - Monitoring?
