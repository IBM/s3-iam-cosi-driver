kubectl delete -f ./examples/bucketaccess.yaml
kubectl delete -f ./examples/bucketaccessclass.yaml
kubectl delete -f ./examples/bucketclaim.yaml
kubectl delete -f ./examples/bucketsecret.yaml
kubectl delete -f ./examples/bucketclass.yaml
kubectl delete -f ./examples/awscliapppod.yaml
kubectl delete configmap setup-aws-script --namespace default