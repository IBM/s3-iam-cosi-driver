# Admin creates these:
kubectl create -f ./examples/simple/bucketsecret.yaml
kubectl create -f ./examples/simple/bucketclass.yaml
kubectl create -f ./examples/simple/bucketaccessclass.yaml

# User creates these:
kubectl create -f ./examples/simple/bucketclaim.yaml
kubectl create -f ./examples/simple/bucketaccess.yaml

# Test Pod
kubectl create configmap setup-aws-script --from-file=examples/setup-aws.sh
kubectl create -f ./examples/simple/awscliapppod.yaml
