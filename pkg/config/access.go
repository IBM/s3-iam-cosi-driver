/*
Copyright 2024-2025 IBM Corporation

Licensed under the MIT License.
*/

package config

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"

	"github.ibm.com/graphene/s3-iam-cosi-driver/pkg/util/s3client"
)

// GetAllowedActions determines the allowed S3 actions based on the access mode
func GetAllowedActions(accessMode string) ([]string, error) {
	var allowedActions []string
	switch accessMode {
	case AccessModeReadOnly:
		allowedActions = s3client.GetActionStrings(s3client.ReadOnlyActions)
	case AccessModeReadWrite:
		allowedActions = s3client.GetActionStrings(s3client.ReadWriteActions)
	case AccessModeWriteOnly:
		allowedActions = s3client.GetActionStrings(s3client.WriteOnlyActions)
	case AccessModeListOnly:
		allowedActions = s3client.GetActionStrings(s3client.ListOnlyActions)
	case AccessModeAdmin:
		allowedActions = s3client.GetActionStrings(s3client.AdminActions)
	default:
		klog.ErrorS(nil, "invalid access mode", "mode", accessMode)
		return nil, status.Error(codes.InvalidArgument, "invalid access mode")
	}
	klog.InfoS("determined allowed actions", "actions", allowedActions)
	return allowedActions, nil
}
