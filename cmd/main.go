/*
Copyright (c) 2024-2025 IBM Corporation

Licensed under the MIT License.
*/

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.ibm.com/graphene/s3-iam-cosi-driver/pkg/config"
	"github.ibm.com/graphene/s3-iam-cosi-driver/pkg/driver"
	"k8s.io/klog/v2"

	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/provisioner"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		klog.InfoS("Signal received", "type", sig)
		cancel()

		<-time.After(30 * time.Second)
		os.Exit(1)
	}()

	if err := run(ctx); err != nil {
		klog.ErrorS(err, "Exiting on error")
	}
}

const driverName = config.DriverName

var (
	driverAddress = flag.String("driver-address", "", "driver address for socket")
)

func init() {
	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "true"); err != nil {
		klog.Exitf("failed to set logtostderr flag: %v", err)
	}
	flag.Parse()

	// Default `driverAddress` based on environment
	if *driverAddress == "" {
		if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
			// Running locally
			*driverAddress = "unix:///tmp/cosi.sock" // Use a local Unix domain socket
			klog.Info("Running locally, using driverAddress: unix:///tmp/cosi.sock")
		} else {
			// Running in-cluster
			*driverAddress = "unix:///var/lib/cosi/cosi.sock"
			klog.Info("Running in-cluster, using driverAddress: unix:///var/lib/cosi/cosi.sock")
		}
	}
}

func run(ctx context.Context) error {
	klog.Info("Driver name: ", driverName)
	identityServer, bucketProvisioner, err := driver.NewDriver(ctx, driverName)
	if err != nil {
		return err
	}

	server, err := provisioner.NewDefaultCOSIProvisionerServer(*driverAddress,
		identityServer,
		bucketProvisioner)
	if err != nil {
		return err
	}

	klog.Info("Starting COSI provisioner server")
	return server.Run(ctx)
}
