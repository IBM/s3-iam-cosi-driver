/*
Copyright 2024-2025 IBM Corporation

Licensed under the MIT License.
*/

package config

// Driver name and annotation constants
const (
	// DriverName is the name of the S3 IAM COSI driver
	DriverName = "s3-iam.objectstorage.k8s.io"
)

// Access mode constants for S3 bucket access
const (
	// AccessModeReadOnly allows read-only access to the bucket
	AccessModeReadOnly = "ro"
	// AccessModeReadWrite allows read and write access to the bucket
	AccessModeReadWrite = "rw"
	// AccessModeWriteOnly allows write-only access to the bucket
	AccessModeWriteOnly = "wo"
	// AccessModeListOnly allows only listing the bucket contents
	AccessModeListOnly = "lo"
	// AccessModeAdmin allows full administrative access to the bucket
	AccessModeAdmin = "admin"
)

// Access mode annotation key
const (
	// AccessModeKey is the key used in annotations to specify the access mode
	AccessModeKey = DriverName + "/access-mode"
)
