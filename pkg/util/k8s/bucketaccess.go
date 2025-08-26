/*
Copyright 2024-2025 IBM Corporation

Licensed under the MIT License.
*/

package k8s

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	objectstoragev1alpha1 "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
	bucketclientset "sigs.k8s.io/container-object-storage-interface-api/client/clientset/versioned"
)

// GetBucketAccessAndClass retrieves the BucketAccess and BucketAccessClass objects
func GetBucketAccessAndClass(ctx context.Context, bucketClientset bucketclientset.Interface, bucketAccessId string) (*objectstoragev1alpha1.BucketAccess, *objectstoragev1alpha1.BucketAccessClass, error) {
	bucketAccess, err := FindBucketAccess(ctx, bucketClientset, bucketAccessId)
	if err != nil {
		klog.ErrorS(err, "failed to find bucket access", "bucketAccessId", bucketAccessId)
		return nil, nil, err
	}

	bucketAccessClassName := bucketAccess.Spec.BucketAccessClassName
	if bucketAccessClassName == "" {
		klog.ErrorS(nil, "bucketAccessClassName is missing from BucketAccess CR")
		return nil, nil, status.Error(codes.InvalidArgument, "bucketAccessClassName is required")
	}

	klog.InfoS("fetching bucket access class", "name", bucketAccessClassName)
	bucketAccessClass, err := bucketClientset.ObjectstorageV1alpha1().BucketAccessClasses().Get(ctx, bucketAccessClassName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to get bucket access class", "name", bucketAccessClassName)
		return nil, nil, status.Error(codes.Internal, "failed to get bucket access class")
	}

	return bucketAccess, bucketAccessClass, nil
}

// FindBucketAccess searches for a BucketAccess CR across all namespaces by its UID
func FindBucketAccess(ctx context.Context, bucketClientset bucketclientset.Interface, bucketAccessId string) (*objectstoragev1alpha1.BucketAccess, error) {
	// List BucketAccess CRs across all namespaces to find the one with matching UID
	bucketAccessList, err := bucketClientset.ObjectstorageV1alpha1().BucketAccesses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to list bucket accesses")
		return nil, status.Error(codes.Internal, "failed to list bucket accesses")
	}

	// Remove "ba-" prefix if it exists
	searchId := strings.TrimPrefix(bucketAccessId, "ba-")

	var bucketAccess *objectstoragev1alpha1.BucketAccess
	for _, ba := range bucketAccessList.Items {
		if string(ba.UID) == searchId {
			bucketAccess = &ba
			break
		}
	}

	if bucketAccess == nil {
		klog.ErrorS(nil, "failed to find bucket access", "uid", searchId)
		return nil, status.Error(codes.NotFound, "bucket access not found")
	}

	return bucketAccess, nil
}
