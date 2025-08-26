/*
Copyright (c) 2024-2025 IBM Corporation

Licensed under the MIT License.
*/

package s3client

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
)

// RawBucketPolicy represents a raw S3 bucket policy
type RawBucketPolicy struct {
	Version   string               `json:"Version"`
	Statement []RawPolicyStatement `json:"Statement"`
}

// RawPolicyStatement represents a raw S3 policy statement
type RawPolicyStatement struct {
	Effect    string      `json:"Effect"`
	Principal interface{} `json:"Principal"`
	Action    interface{} `json:"Action"`
	Resource  []string    `json:"Resource"`
}

// PutRawBucketPolicy applies a raw policy to the bucket
func (s *S3Client) PutRawBucketPolicy(bucket string, policy string) (*s3.PutBucketPolicyOutput, error) {
	confirmRemoveSelfBucketAccess := false // avoids bucket lockout
	p := &s3.PutBucketPolicyInput{
		Bucket:                        &bucket,
		ConfirmRemoveSelfBucketAccess: &confirmRemoveSelfBucketAccess,
		Policy:                        &policy,
	}
	return s.S3.PutBucketPolicy(p)
}

// NewRawDenyAllPolicy creates a raw deny-all policy for a bucket
func NewRawDenyAllPolicy(bucketName string) string {
	policy := RawBucketPolicy{
		Version: "2012-10-17",
		Statement: []RawPolicyStatement{
			{
				Effect:    "Deny",
				Principal: "*",
				Action:    "s3:*",
				Resource: []string{
					fmt.Sprintf("arn:aws:s3:::%s", bucketName),
					fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
				},
			},
		},
	}
	policyJSON, _ := json.MarshalIndent(policy, "", "  ")
	return string(policyJSON)
}

// NewRawAllowAccountPolicy creates a raw policy that allows access for the account user
func NewRawAllowAccountPolicy(bucketName string, accountUser string) string {
	policy := RawBucketPolicy{
		Version: "2012-10-17",
		Statement: []RawPolicyStatement{
			{
				Effect: "Allow",
				Principal: map[string][]string{
					"AWS": {accountUser},
				},
				Action: "s3:*",
				Resource: []string{
					fmt.Sprintf("arn:aws:s3:::%s", bucketName),
					fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
				},
			},
		},
	}
	policyJSON, _ := json.MarshalIndent(policy, "", "  ")
	return string(policyJSON)
}
