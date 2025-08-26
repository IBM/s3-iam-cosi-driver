# Access examples

## Create various BucketAccess types

```sh
kubectl create -f examples/access/ba-admin.yaml
kubectl create -f examples/access/ba-readwrite.yaml
kubectl create -f examples/access/ba-readonly.yaml
kubectl create -f examples/access/ba-writeonly.yaml
kubectl create -f examples/access/ba-listonly.yaml
```

Verify they were created:

```sh
kubectl get bucketaccess -n default
```

## Setup test pod

1. Create ConfigMap with script to setup AWS Client

```sh
kubectl create configmap setup-aws-script -n default --from-file=examples/access/setup-aws.sh
```

1. Create Pod

```sh
kubectl create -f ./examples/simple/awscliapppod-ro.yaml
kubectl create -f ./examples/simple/awscliapppod-rw.yaml
```
