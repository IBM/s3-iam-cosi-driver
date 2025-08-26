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

package s3client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	rgwRegion   = "us-east-1" // default region
	HttpTimeOut = 15 * time.Second
)

// S3Client wraps the S3 and IAM APIs
type S3Client struct {
	S3  s3iface.S3API
	IAM IAMClientInterface
}

func NewS3Client(params *S3ClientParams, debug bool) (*S3Client, error) {
	logLevel := aws.LogOff
	if debug {
		logLevel = aws.LogDebug
	}

	// Create base HTTP client
	client := http.Client{
		Timeout: HttpTimeOut,
	}

	// Handle TLS configuration
	klog.V(5).InfoS("Creating S3 client with endpoint", "endpoint", params.Endpoint)
	tlsEnabled := false
	insecure := true
	if strings.HasPrefix(params.Endpoint, "https") {
		klog.V(5).InfoS("Using secure endpoint")
		insecure = false
		if len(params.TlsCert) > 0 {
			klog.V(5).InfoS("Using TLS certificate")
			tlsEnabled = true
			client.Transport = buildTransportTLS(params.TlsCert, insecure)
		}
	} else {
		klog.V(5).InfoS("Using insecure endpoint", "endpoint", params.Endpoint)
	}

	// Create S3 session with S3 endpoint
	s3Session, err := session.NewSession(
		aws.NewConfig().
			WithRegion(params.Region).
			WithCredentials(credentials.NewStaticCredentials(params.AccessKey, params.SecretKey, "")).
			WithEndpoint(params.GetFullEndpoint()).
			WithS3ForcePathStyle(true).
			WithMaxRetries(5).
			WithDisableSSL(!tlsEnabled).
			WithHTTPClient(&client).
			WithLogLevel(logLevel),
	)
	if err != nil {
		return nil, err
	}
	s3Svc := s3.New(s3Session)

	// Create IAM client with IAM endpoint
	iamClient, err := NewIAMClient(params, debug)
	if err != nil {
		return nil, err
	}

	return &S3Client{
		S3:  s3Svc,
		IAM: iamClient,
	}, nil
}

// CreateBucket creates a bucket with the given name
func (s *S3Client) CreateBucketNoInfoLogging(name string) error {
	return s.createBucket(name, false)
}

// CreateBucket creates a bucket with the given name
func (s *S3Client) CreateBucket(name string) error {
	return s.createBucket(name, true)
}

func (s *S3Client) createBucket(name string, infoLogging bool) error {
	if infoLogging {
		klog.InfoS("creating bucket", "name", name)
	} else {
		klog.InfoS("creating bucket", "name", name)
	}
	bucketInput := &s3.CreateBucketInput{
		Bucket: &name,
	}
	_, err := s.S3.CreateBucket(bucketInput)
	if err != nil {
		return err
	}

	if infoLogging {
		klog.InfoS("Successfully created bucket", "name", name)
	} else {
		klog.InfoS("Successfully created bucket", "name", name)
	}
	return nil
}

// DeleteBucket function deletes given bucket using s3 client
func (s *S3Client) DeleteBucket(name string) (bool, error) {
	_, err := s.S3.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		klog.ErrorS(err, "Failed to delete bucket")
		return false, err

	}
	return true, nil
}

// PutObjectInBucket function puts an object in a bucket using s3 client
func (s *S3Client) PutObjectInBucket(bucketname string, body string, key string,
	contentType string) (bool, error) {
	_, err := s.S3.PutObject(&s3.PutObjectInput{
		Body:        strings.NewReader(body),
		Bucket:      &bucketname,
		Key:         &key,
		ContentType: &contentType,
	})
	if err != nil {
		klog.ErrorS(err, "Failed to put object in bucket")
		return false, err

	}
	return true, nil
}

// GetObjectInBucket function retrieves an object from a bucket using s3 client
func (s *S3Client) GetObjectInBucket(bucketname string, key string) (string, error) {
	result, err := s.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	})

	if err != nil {
		klog.ErrorS(err, "Failed to retrieve object from bucket")
		return "ERROR_ OBJECT NOT FOUND", err

	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// DeleteObjectInBucket function deletes given bucket using s3 client
func (s *S3Client) DeleteObjectInBucket(bucketname string, key string) (bool, error) {
	_, err := s.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return true, nil
			case s3.ErrCodeNoSuchKey:
				return true, nil
			}
		}
		klog.ErrorS(err, "Failed to delete object from bucket")
		return false, err

	}
	return true, nil
}

// PutBucketPolicy applies the policy to the bucket
func (s *S3Client) PutBucketPolicy(bucket string, policy BucketPolicy) (*s3.PutBucketPolicyOutput, error) {

	confirmRemoveSelfBucketAccess := false
	serializedPolicy, _ := json.Marshal(policy)
	consumablePolicy := string(serializedPolicy)

	p := &s3.PutBucketPolicyInput{
		Bucket:                        &bucket,
		ConfirmRemoveSelfBucketAccess: &confirmRemoveSelfBucketAccess,
		Policy:                        &consumablePolicy,
	}
	out, err := s.S3.PutBucketPolicy(p)
	if err != nil {
		return out, err
	}
	return out, nil
}

func (s *S3Client) GetBucketPolicy(bucket string) (*BucketPolicy, error) {
	out, err := s.S3.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: &bucket,
	})
	if err != nil {
		return nil, err
	}

	policy := &BucketPolicy{}
	err = json.Unmarshal([]byte(*out.Policy), policy)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

func buildTransportTLS(tlsCert []byte, insecure bool) *http.Transport {
	//nolint:gosec // is enabled only for testing
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: insecure,
	}

	if len(tlsCert) > 0 {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(tlsCert)
		tlsConfig.RootCAs = caCertPool
	}

	return &http.Transport{
		TLSClientConfig: tlsConfig,
	}
}

// InitializeClients creates and returns an S3 client using the provided parameters
func InitializeClients(ctx context.Context, clientset *kubernetes.Clientset, parameters map[string]string) (*S3Client, error) {
	klog.V(5).Infof("Initializing clients %v", parameters)

	accountSecretName, namespace, err := FetchSecretNameAndNamespace(parameters)
	if err != nil {
		return nil, err
	}

	accountSecret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, accountSecretName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to get object store user secret")
		return nil, status.Error(codes.Internal, "Failed to get object store user secret")
	}

	s3Params, err := FetchParameters(accountSecret.Data)
	if err != nil {
		return nil, err
	}

	s3Client, err := NewS3Client(s3Params, false)
	if err != nil {
		klog.ErrorS(err, "Failed to create s3 client")
		return nil, status.Error(codes.Internal, "Failed to create s3 client")
	}
	return s3Client, nil
}

// FetchSecretNameAndNamespace retrieves the secret name and namespace from the parameters
func FetchSecretNameAndNamespace(parameters map[string]string) (string, string, error) {
	secretName := parameters["accountSecret"]
	namespace := os.Getenv("POD_NAMESPACE")
	if parameters["accountSecretNamespace"] != "" {
		namespace = parameters["accountSecretNamespace"]
	}
	if secretName == "" || namespace == "" {
		return "", "", status.Error(codes.InvalidArgument, "accountSecret and accountSecretNamespace is required")
	}

	return secretName, namespace, nil
}

// FetchParameters retrieves and validates S3 client parameters from secret data
func FetchParameters(secretData map[string][]byte) (*S3ClientParams, error) {
	endPoint := string(secretData["Endpoint"])
	s3Port := string(secretData["S3Port"])
	iamPort := string(secretData["IAMPort"])
	accountName := string(secretData["AccountName"])
	accessKey := string(secretData["AccessKey"])
	secretKey := string(secretData["SecretKey"])
	region := string(secretData["Region"])

	// Handle TlsCert - use the raw bytes, will be empty slice if not present
	tlsCert := secretData["TlsCert"]

	if endPoint == "" || accessKey == "" || secretKey == "" {
		return nil, status.Error(codes.InvalidArgument, "endpoint, accessKeyID and secretKey are required")
	}

	if s3Port == "" || iamPort == "" {
		return nil, status.Error(codes.InvalidArgument, "s3Port and iamPort are required")
	}

	if accountName == "" {
		return nil, status.Error(codes.InvalidArgument, "accountName is required")
	}

	// Validate that the endpoint includes a protocol
	if !strings.HasPrefix(endPoint, "http://") && !strings.HasPrefix(endPoint, "https://") {
		return nil, status.Error(codes.InvalidArgument, "endpoint must include http:// or https:// protocol")
	}

	// AWS requires a region
	if region == "" {
		klog.Warning("region is not set, using default region", "region", rgwRegion)
		region = rgwRegion
	}

	return &S3ClientParams{
		Endpoint:    endPoint,
		S3Port:      s3Port,
		IAMPort:     iamPort,
		AccountName: accountName,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		TlsCert:     tlsCert,
		Region:      region,
	}, nil
}

// AddUserToBucketPolicy adds a user to a bucket's policy with the specified access mode
func (s *S3Client) AddUserToBucketPolicy(bucketName, userName string, allowedActions []string) error {
	klog.InfoS("Attempting to add user to bucket policy",
		"bucketName", bucketName,
		"username", userName,
		"actions", allowedActions)

	// Get current policy
	klog.V(5).InfoS("getting current bucket policy", "bucketName", bucketName)
	policy, err := s.S3.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() != "NoSuchBucketPolicy" {
			klog.ErrorS(err, "Failed to get bucket policy",
				"bucketName", bucketName,
				"errorCode", aerr.Code())
			return err
		}
		// No policy exists yet, that's fine
		klog.InfoS("No existing bucket policy found, will create new one", "bucketName", bucketName)
		policy = nil
	}

	// Get user ID
	userOutput, err := s.IAM.GetUser(userName)
	if err != nil {
		klog.ErrorS(err, "Failed to get user ID",
			"bucketName", bucketName,
			"username", userName)
		return err
	}

	// Create new statement for the user
	newStatement := RawPolicyStatement{
		Effect: "Allow",
		Principal: map[string][]string{
			"AWS": {*userOutput.User.UserId},
		},
		Action: allowedActions,
		Resource: []string{
			fmt.Sprintf("arn:aws:s3:::%s", bucketName),
			fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
		},
	}

	var policyJSON []byte
	if policy != nil {
		// Parse existing policy
		klog.V(5).InfoS("Parsing existing policy", "bucketName", bucketName)
		var rawPolicy RawBucketPolicy
		err = json.Unmarshal([]byte(*policy.Policy), &rawPolicy)
		if err != nil {
			klog.ErrorS(err, "Failed to unmarshal existing policy",
				"bucketName", bucketName,
				"policy", *policy.Policy)
			return err
		}

		klog.V(5).InfoS("Existing policy",
			"bucketName", bucketName,
			"statements", len(rawPolicy.Statement))

		// Check if user already has access
		for _, stmt := range rawPolicy.Statement {
			if principal, ok := stmt.Principal.(map[string]interface{}); ok {
				if users, ok := principal["AWS"].([]interface{}); ok {
					for _, u := range users {
						if u.(string) == userName {
							klog.InfoS("User already has access to bucket",
								"bucketName", bucketName,
								"username", userName)
							return nil
						}
					}
				}
			}
		}

		// Append new statement
		klog.V(5).InfoS("appending new statement to existing policy", "bucketName", bucketName)
		rawPolicy.Statement = append(rawPolicy.Statement, newStatement)
		policyJSON, err = json.MarshalIndent(rawPolicy, "", "  ")
	} else {
		// Create new policy with just this statement
		klog.V(5).InfoS("creating new policy", "bucketName", bucketName)
		rawPolicy := RawBucketPolicy{
			Version:   "2012-10-17",
			Statement: []RawPolicyStatement{newStatement},
		}
		policyJSON, err = json.MarshalIndent(rawPolicy, "", "  ")
	}

	if err != nil {
		klog.ErrorS(err, "Failed to marshal policy",
			"bucketName", bucketName)
		return err
	}

	klog.V(5).InfoS("Setting bucket policy",
		"bucketName", bucketName,
		"policy", string(policyJSON))

	// Put the updated policy
	policyStr := string(policyJSON)
	_, err = s.S3.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: &policyStr,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			klog.ErrorS(err, "Failed to set bucket policy",
				"bucketName", bucketName,
				"errorCode", aerr.Code(),
				"errorMessage", aerr.Message())
		} else {
			klog.ErrorS(err, "Failed to set bucket policy",
				"bucketName", bucketName)
		}
		return err
	}

	klog.InfoS("Successfully added user to bucket policy",
		"bucketName", bucketName,
		"username", userName,
		"actions", allowedActions)
	return nil
}

// RemoveUserFromBucketPolicy removes a user from a bucket's policy
func (s *S3Client) RemoveUserFromBucketPolicy(bucketName, userName string) error {
	klog.InfoS("Attempting to remove user from bucket policy",
		"bucketName", bucketName,
		"username", userName)

	// Get the current bucket policy
	policy, err := s.S3.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NoSuchBucketPolicy" {
			// No policy exists, nothing to remove
			klog.InfoS("no bucket policy exists, nothing to remove",
				"bucketName", bucketName)
			return nil
		}
		klog.ErrorS(err, "Failed to get bucket policy",
			"bucketName", bucketName)
		return err
	}

	klog.InfoS("current bucket policy",
		"bucketName", bucketName,
		"policy", *policy.Policy)

	// Parse the policy JSON
	var rawPolicy RawBucketPolicy
	err = json.Unmarshal([]byte(*policy.Policy), &rawPolicy)
	if err != nil {
		klog.ErrorS(err, "failed to unmarshal bucket policy",
			"bucketName", bucketName)
		return err
	}

	klog.InfoS("parsed bucket policy",
		"bucketName", bucketName,
		"statements", len(rawPolicy.Statement))

	// Get user ID
	userOutput, err := s.IAM.GetUser(userName)
	if err != nil {
		klog.ErrorS(err, "Failed to get user ID",
			"bucketName", bucketName,
			"username", userName)
		return err
	}
	klog.V(5).InfoS("user ID",
		"bucketName", bucketName,
		"username", userName,
		"userId", *userOutput.User.UserId)

	// Filter out statements that grant access to this user
	var newStatements []RawPolicyStatement
	for i, stmt := range rawPolicy.Statement {
		// Skip statements that grant access to this user
		if principal, ok := stmt.Principal.(map[string]interface{}); ok {
			if users, ok := principal["AWS"].([]interface{}); ok {
				// Check if this statement affects our user
				affectsUser := false
				for _, u := range users {
					if u.(string) == *userOutput.User.UserId {
						affectsUser = true
						break
					}
				}
				if affectsUser {
					klog.InfoS("removing statement that affects user",
						"bucketName", bucketName,
						"username", userName,
						"userId", *userOutput.User.UserId,
						"statementIndex", i)
					continue // Skip this statement
				}
			}
		}
		newStatements = append(newStatements, stmt)
	}

	klog.InfoS("filtered policy statements",
		"bucketName", bucketName,
		"originalCount", len(rawPolicy.Statement),
		"newCount", len(newStatements))

	// If we removed all statements, delete the entire policy
	if len(newStatements) == 0 {
		klog.InfoS("all statements removed, deleting entire policy",
			"bucketName", bucketName)
		_, err = s.S3.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			klog.ErrorS(err, "failed to delete bucket policy",
				"bucketName", bucketName)
			return err
		}
		klog.InfoS("deleted empty bucket policy",
			"bucketName", bucketName)
		return nil
	}

	// Update the policy with the filtered statements
	rawPolicy.Statement = newStatements
	policyJSON, err := json.MarshalIndent(rawPolicy, "", "  ")
	if err != nil {
		klog.ErrorS(err, "failed to marshal updated policy",
			"bucketName", bucketName)
		return err
	}

	klog.InfoS("updated policy content",
		"bucketName", bucketName,
		"policy", string(policyJSON))

	// Put the updated policy back
	policyStr := string(policyJSON)
	_, err = s.S3.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: &policyStr,
	})
	if err != nil {
		klog.ErrorS(err, "failed to update bucket policy",
			"bucketName", bucketName)
		return err
	}

	klog.InfoS("successfully removed user from bucket policy",
		"bucketName", bucketName,
		"username", userName)
	return nil
}
