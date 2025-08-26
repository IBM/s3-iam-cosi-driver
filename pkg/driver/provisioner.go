/*
Copyright 2021 The Ceph-COSI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Modifications Copyright (c) 2024-2025 IBM Corporation

Licensed under the MIT License.
*/

package driver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.ibm.com/graphene/s3-iam-cosi-driver/pkg/config"
	"github.ibm.com/graphene/s3-iam-cosi-driver/pkg/util/k8s"
	"github.ibm.com/graphene/s3-iam-cosi-driver/pkg/util/s3client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	bucketclientset "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned"
	cosispec "sigs.k8s.io/container-object-storage-interface-spec"
)

// contains two clients
// 1.) for AdminOps : mainly for user related operations
// 2.) for S3 operations : mainly for bucket related operations
type provisionerServer struct {
	Provisioner     string
	Clientset       *kubernetes.Clientset
	KubeConfig      *rest.Config
	BucketClientset bucketclientset.Interface
}

var _ cosispec.ProvisionerServer = &provisionerServer{}

func NewProvisionerServer(provisioner string) (cosispec.ProvisionerServer, error) {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		kubeConfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
		klog.Info("Running with local kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	bucketClientset, err := bucketclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return &provisionerServer{
		Provisioner:     provisioner,
		Clientset:       clientset,
		KubeConfig:      kubeConfig,
		BucketClientset: bucketClientset,
	}, nil
}

// ProvisionerCreateBucket is an idempotent method for creating buckets
// It is expected to create the same bucket given a bucketName and protocol
// If the bucket already exists, then it MUST return codes.AlreadyExists
// Return values
//
//	nil -                   Bucket successfully created
//	codes.AlreadyExists -   Bucket already exists. No more retries
//	non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (s *provisionerServer) DriverCreateBucket(ctx context.Context,
	req *cosispec.DriverCreateBucketRequest) (*cosispec.DriverCreateBucketResponse, error) {
	klog.Infof("req %v", req)

	bucketName := req.GetName()
	klog.InfoS("Creating Bucket", "name", bucketName)

	parameters := req.GetParameters()

	s3Client, err := s3client.InitializeClients(ctx, s.Clientset, parameters)
	if err != nil {
		klog.ErrorS(err, "Failed to initialize clients")
		return nil, status.Error(codes.Internal, "Failed to initialize clients")
	}

	err = s3Client.CreateBucket(bucketName)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			klog.InfoS("DEBUG: after s3 call", "ok", ok, "aerr", aerr)
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				klog.InfoS("Bucket already exists", "name", bucketName)
				return &cosispec.DriverCreateBucketResponse{
					BucketId: bucketName,
				}, nil
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				klog.InfoS("Bucket already owned by you", "name", bucketName)
				return &cosispec.DriverCreateBucketResponse{
					BucketId: bucketName,
				}, nil
			}
		}
		klog.ErrorS(err, "Failed to create bucket", "bucketName", bucketName)
		return nil, status.Error(codes.Internal, "Failed to create bucket")
	}

	// TODO: this broke recently on Scale S3 (Noobaa).  It now self-locks the bucket.
	// We need to figure out how to set the policy correctly.
	/*
		// Set initial bucket policy to deny all actions
		// Create a simple deny-all policy that matches the exact format
		policyJSON := s3client.NewRawDenyAllPolicy(bucketName)

		// Log the policy for debugging
		klog.InfoS("Setting bucket policy", "policy", string(policyJSON))

		// Try using the policy
		_, err = s3Client.PutRawBucketPolicy(bucketName, policyJSON)
		if err != nil {
			klog.ErrorS(err, "failed to set initial deny-all policy")
			// Don't return error here, as the bucket was created successfully
		} else {
			klog.InfoS("Successfully set initial deny-all policy", "bucketName", bucketName)
		}
	*/

	klog.InfoS("Successfully created Backend Bucket", "bucketName", bucketName)

	return &cosispec.DriverCreateBucketResponse{
		BucketId: bucketName,
	}, nil
}

func (s *provisionerServer) DriverDeleteBucket(ctx context.Context,
	req *cosispec.DriverDeleteBucketRequest) (*cosispec.DriverDeleteBucketResponse, error) {
	klog.V(5).Infof("req %v", req)
	bucketName := req.GetBucketId()
	klog.V(3).InfoS("Deleting Bucket", "name", bucketName)
	bucket, err := s.BucketClientset.ObjectstorageV1alpha1().Buckets().Get(ctx, bucketName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get bucket", "bucketName", bucketName)
		return nil, status.Error(codes.Internal, "failed to get bucket")
	}

	parameters := bucket.Spec.Parameters
	s3Client, err := s3client.InitializeClients(ctx, s.Clientset, parameters)
	if err != nil {
		klog.ErrorS(err, "failed to initialize clients")
		return nil, status.Error(codes.Internal, "failed to initialize clients")
	}

	_, err = s3Client.DeleteBucket(bucketName)
	if err != nil {
		klog.ErrorS(err, "failed to delete bucket", "bucketName", bucketName)
		return nil, status.Error(codes.Internal, "failed to delete bucket")
	}
	klog.InfoS("Successfully deleted Backend Bucket", "bucketName", bucketName)
	return &cosispec.DriverDeleteBucketResponse{}, nil
}

func (s *provisionerServer) DriverGrantBucketAccess(ctx context.Context,
	req *cosispec.DriverGrantBucketAccessRequest) (*cosispec.DriverGrantBucketAccessResponse, error) {
	klog.Infof("req %v", req)
	bucketName := req.GetBucketId()
	bucketAccessId := req.GetName()

	// Generate username with format: cosi-user-<bucketclaim>-<random>
	userName := fmt.Sprintf("cosi-user-%s", bucketAccessId)
	klog.Info("Granting user accessPolicy to bucket ", "userName: ", userName, "bucketName: ", bucketName)

	// Get parameters and initialize S3 client
	parameters := req.GetParameters()
	accountSecretName, namespace, err := s3client.FetchSecretNameAndNamespace(parameters)
	if err != nil {
		return nil, err
	}

	accountSecret, err := s.Clientset.CoreV1().Secrets(namespace).Get(ctx, accountSecretName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get CES account secret")
		return nil, status.Error(codes.Internal, "failed to get ces account secret")
	}

	s3Params, err := s3client.FetchParameters(accountSecret.Data)
	if err != nil {
		klog.ErrorS(err, "failed to fetch S3 parameters from secret")
		return nil, err
	}

	s3Client, err := s3client.InitializeClients(ctx, s.Clientset, parameters)
	if err != nil {
		klog.ErrorS(err, "failed to initialize clients")
		return nil, status.Error(codes.Internal, "failed to initialize clients")
	}

	// Get bucket access and class information
	_, bucketAccessClass, err := k8s.GetBucketAccessAndClass(ctx, s.BucketClientset, bucketAccessId)
	if err != nil {
		return nil, err
	}

	// Get access mode and determine allowed actions
	accessMode := bucketAccessClass.Annotations[config.AccessModeKey]
	if accessMode == "" {
		// Default to admin mode if access mode is not specified
		accessMode = config.AccessModeAdmin
	}

	allowedActions, err := config.GetAllowedActions(accessMode)
	if err != nil {
		return nil, err
	}

	// Create or get IAM user and access key
	accessKey, err := s3Client.EnsureIAMUser(ctx, userName)
	if err != nil {
		return nil, err
	}

	// Add user to bucket policy
	klog.InfoS("adding user to bucket policy", "bucketName", bucketName, "userName", userName, "actions", allowedActions)
	err = s3Client.AddUserToBucketPolicy(bucketName, userName, allowedActions)
	if err != nil {
		klog.ErrorS(err, "failed to add user to bucket policy", "bucketName", bucketName, "userName", userName)
		return nil, status.Error(codes.Internal, "failed to add user to bucket policy")
	}

	klog.InfoS("Successfully granted bucket access", "bucketName", bucketName, "accessMode", accessMode)
	return &cosispec.DriverGrantBucketAccessResponse{
		AccountId: userName,
		Credentials: fetchUserCredentials(
			*accessKey.AccessKeyId,
			*accessKey.SecretAccessKey,
			s3Params.GetFullEndpoint(),
			"",
		),
	}, nil
}

func (s *provisionerServer) DriverRevokeBucketAccess(ctx context.Context,
	req *cosispec.DriverRevokeBucketAccessRequest) (*cosispec.DriverRevokeBucketAccessResponse, error) {
	klog.Infof("req %v", req)

	userName := req.GetAccountId()
	bucketName := req.GetBucketId()
	klog.InfoS("Revoking user accessPolicy from bucket",
		"userName", userName,
		"bucketName", bucketName)

	// Get the bucket to find the bucket claim name
	bucket, err := s.BucketClientset.ObjectstorageV1alpha1().Buckets().Get(ctx, bucketName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get bucket", "bucketName", bucketName)
		return nil, status.Error(codes.Internal, "failed to get bucket")
	}

	parameters := bucket.Spec.Parameters
	accountSecretName, namespace, err := s3client.FetchSecretNameAndNamespace(parameters)
	if err != nil {
		return nil, err
	}

	accountSecret, err := s.Clientset.CoreV1().Secrets(namespace).Get(ctx, accountSecretName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get CES account secret")
		return nil, status.Error(codes.Internal, "failed to get ces account secret")
	}

	_, err = s3client.FetchParameters(accountSecret.Data)
	if err != nil {
		return nil, err
	}

	s3Client, err := s3client.InitializeClients(ctx, s.Clientset, parameters)
	if err != nil {
		klog.ErrorS(err, "failed to initialize clients")
		return nil, status.Error(codes.Internal, "failed to initialize clients")
	}
	// Remove user from bucket policy
	err = s3Client.RemoveUserFromBucketPolicy(bucketName, userName)
	if err != nil {
		klog.ErrorS(err, "failed to remove user from bucket policy",
			"userName", userName,
			"bucketName", bucketName)
		return nil, status.Error(codes.Internal, "failed to remove user from bucket policy")
	}

	// Delete the IAM user
	err = s3Client.IAM.DeleteUser(userName)
	if err != nil {
		klog.ErrorS(err, "failed to delete IAM user",
			"userName", userName)
		return nil, status.Error(codes.Internal, "failed to delete IAM user")
	}

	return &cosispec.DriverRevokeBucketAccessResponse{}, nil
}

func fetchUserCredentials(accessKey, secretKey, endpoint string, region string) map[string]*cosispec.CredentialDetails {
	s3Keys := make(map[string]string)
	s3Keys["accessKeyID"] = accessKey
	s3Keys["accessSecretKey"] = secretKey
	s3Keys["endpoint"] = endpoint
	s3Keys["region"] = region
	creds := &cosispec.CredentialDetails{
		Secrets: s3Keys,
	}
	credDetails := make(map[string]*cosispec.CredentialDetails)
	credDetails["s3"] = creds
	return credDetails
}
