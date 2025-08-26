/*
Copyright (c) 2024-2025 IBM Corporation

Licensed under the MIT License.
*/

package s3client

import (
	"regexp"
	"strings"

	"k8s.io/klog/v2"
)

// S3ClientParams holds the parameters for creating an S3 client
type S3ClientParams struct {
	Endpoint    string
	S3Port      string
	IAMPort     string
	AccountName string
	AccessKey   string
	SecretKey   string
	TlsCert     []byte
	Region      string
}

// GetFullEndpoint returns the complete endpoint URL with port if needed
func (p *S3ClientParams) GetFullEndpoint() string {
	// Check if the endpoint already contains a port (contains a colon followed by digits)
	if strings.Contains(p.Endpoint, ":") && regexp.MustCompile(`:\d+`).MatchString(p.Endpoint) {
		return p.Endpoint
	}

	// If S3Port is empty, return the endpoint as is (will use default port)
	if p.S3Port == "" {
		return p.Endpoint
	}

	// Extract the protocol (http:// or https://)
	protocol := ""
	finalEndpoint := p.Endpoint
	if strings.HasPrefix(p.Endpoint, "http://") {
		protocol = "http://"
		finalEndpoint = strings.TrimPrefix(p.Endpoint, "http://")
	} else if strings.HasPrefix(p.Endpoint, "https://") {
		protocol = "https://"
		finalEndpoint = strings.TrimPrefix(p.Endpoint, "https://")
	}

	// Add the port to the endpoint
	return protocol + finalEndpoint + ":" + p.S3Port
}

// GetFullIAMEndpoint returns the complete IAM endpoint URL with port if needed
func (p *S3ClientParams) GetFullIAMEndpoint() string {
	// Extract the protocol (http:// or https://)
	protocol := "https://"
	endpoint := p.Endpoint
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	}

	klog.V(5).InfoS("IAM endpoint construction",
		"originalEndpoint", p.Endpoint,
		"strippedEndpoint", endpoint,
		"protocol", protocol,
		"iamPort", p.IAMPort)

	// Check if the endpoint already contains a port
	if strings.Contains(endpoint, ":") && regexp.MustCompile(`:\d+`).MatchString(endpoint) {
		finalEndpoint := protocol + endpoint
		klog.V(5).InfoS("IAM endpoint with existing port", "finalEndpoint", finalEndpoint)
		return finalEndpoint
	}

	// If IAMPort is empty, return the endpoint as is (will use default port)
	if p.IAMPort == "" {
		finalEndpoint := protocol + endpoint
		klog.V(5).InfoS("IAM endpoint without port", "finalEndpoint", finalEndpoint)
		return finalEndpoint
	}

	// Add the port to the endpoint
	finalEndpoint := protocol + endpoint + ":" + p.IAMPort
	klog.V(5).InfoS("IAM endpoint with added port", "finalEndpoint", finalEndpoint)
	return finalEndpoint
}
