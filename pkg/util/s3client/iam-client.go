/*
Copyright (c) 2024-2025 IBM Corporation

Licensed under the MIT License.
*/

package s3client

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// IAMClientInterface is an interface for IAM operations
type IAMClientInterface interface {
	GetUser(userName string) (*iam.GetUserOutput, error)
	CreateUser(userName string) (*iam.CreateUserOutput, error)
	DeleteUser(userName string) error
	CreateAccessKey(userName string) (*iam.CreateAccessKeyOutput, error)
	ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error)
	GetAccessKeyLastUsed(input *iam.GetAccessKeyLastUsedInput) (*iam.GetAccessKeyLastUsedOutput, error)
}

// IAMClient wraps the IAM API
type IAMClient struct {
	api *iam.IAM
}

// NewIAMClient creates a new IAM client
func NewIAMClient(params *S3ClientParams, debug bool) (*IAMClient, error) {
	logLevel := aws.LogOff
	if debug {
		logLevel = aws.LogDebug
	}

	// Create base HTTP client
	client := http.Client{
		Timeout: HttpTimeOut,
	}

	// Handle TLS configuration
	klog.V(5).InfoS("Creating IAM client with endpoint", "endpoint", params.Endpoint)
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

	// Create IAM session with IAM endpoint
	iamSession, err := session.NewSession(
		aws.NewConfig().
			WithRegion("us-east-1").
			WithCredentials(credentials.NewStaticCredentials(params.AccessKey, params.SecretKey, "")).
			WithEndpoint(params.GetFullIAMEndpoint()).
			WithMaxRetries(5).
			WithDisableSSL(!tlsEnabled).
			WithHTTPClient(&client).
			WithLogLevel(logLevel).
			WithS3ForcePathStyle(true),
	)
	if err != nil {
		return nil, err
	}

	return &IAMClient{
		api: iam.New(iamSession),
	}, nil
}

// GetUser gets a user by name
func (a *IAMClient) GetUser(userName string) (*iam.GetUserOutput, error) {
	input := &iam.GetUserInput{
		UserName: aws.String(userName),
	}
	return a.api.GetUser(input)
}

// CreateUser creates a new IAM user
func (a *IAMClient) CreateUser(userName string) (*iam.CreateUserOutput, error) {
	input := &iam.CreateUserInput{
		UserName: aws.String(userName),
	}
	return a.api.CreateUser(input)
}

// CreateAccessKey creates a new access key for a user
func (a *IAMClient) CreateAccessKey(userName string) (*iam.CreateAccessKeyOutput, error) {
	input := &iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	}
	return a.api.CreateAccessKey(input)
}

// ListAccessKeys lists access keys for a user
func (a *IAMClient) ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {
	return a.api.ListAccessKeys(input)
}

// GetAccessKeyLastUsed gets the last used information for an access key
func (a *IAMClient) GetAccessKeyLastUsed(input *iam.GetAccessKeyLastUsedInput) (*iam.GetAccessKeyLastUsedOutput, error) {
	return a.api.GetAccessKeyLastUsed(input)
}

// DeleteAccessKey deletes an access key for a user
func (a *IAMClient) DeleteAccessKey(input *iam.DeleteAccessKeyInput) (*iam.DeleteAccessKeyOutput, error) {
	return a.api.DeleteAccessKey(input)
}

// DeleteUser deletes an IAM user
func (a *IAMClient) DeleteUser(userName string) error {
	klog.InfoS("Attempting to delete IAM user", "username", userName)

	// First check if the user exists
	_, err := a.GetUser(userName)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NoSuchEntity" {
			// User doesn't exist, nothing to delete
			klog.InfoS("User does not exist, nothing to delete", "username", userName)
			return nil
		}
		// Some other error occurred
		klog.ErrorS(err, "failed to check if user exists", "username", userName)
		return err
	}

	// First, delete all access keys for the user
	keys, err := a.api.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		klog.ErrorS(err, "failed to list access keys for user", "username", userName)
		return err
	}

	for _, key := range keys.AccessKeyMetadata {
		_, err := a.api.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			UserName:    aws.String(userName),
			AccessKeyId: key.AccessKeyId,
		})
		if err != nil {
			klog.ErrorS(err, "Failed to delete access key",
				"username", userName,
				"accessKeyId", aws.StringValue(key.AccessKeyId))
			return err
		}
	}

	// Now delete the user
	_, err = a.api.DeleteUser(&iam.DeleteUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			klog.ErrorS(err, "DeleteUser AWS error details",
				"username", userName,
				"errorCode", aerr.Code(),
				"errorMessage", aerr.Message(),
				"originalError", aerr.OrigErr())
		} else {
			klog.ErrorS(err, "DeleteUser non-AWS error",
				"username", userName)
		}
		return err
	}

	klog.InfoS("Successfully deleted IAM user", "username", userName)
	return nil
}

// EnsureIAMUser ensures the IAM user exists and returns its access key.
// If the user exists, it will return an existing access key or create a new one if none exist.
func (s *S3Client) EnsureIAMUser(ctx context.Context, userName string) (*iam.AccessKey, error) {
	klog.InfoS("Checking if user exists before creation", "userName", userName)
	_, err := s.IAM.GetUser(userName)
	if err != nil {
		klog.InfoS("User does not exist, attempting to create", "userName", userName, "error", err)
		_, err = s.IAM.CreateUser(userName)
		if err != nil {
			klog.ErrorS(err, "Failed to create IAM user", "userName", userName)
			return nil, status.Error(codes.Internal, "Failed to create IAM user")
		}
		klog.InfoS("Successfully created IAM user", "userName", userName)
	} else {
		klog.InfoS("User already exists, checking for existing access keys", "userName", userName)
		// List existing access keys
		listKeysInput := &iam.ListAccessKeysInput{
			UserName: aws.String(userName),
		}
		listKeysOutput, err := s.IAM.ListAccessKeys(listKeysInput)
		if err != nil {
			klog.ErrorS(err, "Failed to list access keys", "userName", userName)
			return nil, status.Error(codes.Internal, "Failed to list access keys")
		}

		// If user has less than 2 access keys, we can create a new one
		if len(listKeysOutput.AccessKeyMetadata) < 2 {
			klog.InfoS("User has less than 2 access keys, creating new one", "userName", userName, "existingKeys", len(listKeysOutput.AccessKeyMetadata))
		} else {
			// User already has 2 access keys (AWS limit), return the most recently created one
			var latestKey *iam.AccessKeyMetadata
			for _, key := range listKeysOutput.AccessKeyMetadata {
				if latestKey == nil || key.CreateDate.After(*latestKey.CreateDate) {
					latestKey = key
				}
			}
			if latestKey != nil {
				klog.InfoS("Using existing access key", "userName", userName, "accessKeyId", *latestKey.AccessKeyId)
				// Create a new access key to get the secret
				createKeyOutput, err := s.IAM.CreateAccessKey(userName)
				if err != nil {
					klog.ErrorS(err, "Failed to create new access key", "userName", userName)
					return nil, status.Error(codes.Internal, "Failed to create new access key")
				}
				return createKeyOutput.AccessKey, nil
			}
		}
	}

	klog.InfoS("Creating access keys for user", "userName", userName)
	accessKeyResult, err := s.IAM.CreateAccessKey(userName)
	if err != nil {
		klog.ErrorS(err, "Failed to create access key", "userName", userName)
		return nil, status.Error(codes.Internal, "Failed to create access key")
	}

	return accessKeyResult.AccessKey, nil
}
