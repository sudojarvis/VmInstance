package main

import (
	"context"
	"fmt"
	"io"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

func createInstanceWithFirewall(w io.Writer, projectID, zone, instanceName, machineType, sourceImage, networkName, firewallRuleName string) error {

	ctx := context.Background()

	// Create a compute client
	computeClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %w", err)
	}
	defer computeClient.Close()

	// Create a firewall client
	firewallClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewFirewallsRESTClient: %w", err)
	}
	defer firewallClient.Close()

	// Define and create the firewall rule
	firewallRule := &computepb.Firewall{
		Allowed: []*computepb.Allowed{
			{
				IPProtocol: proto.String("tcp"),
				Ports:      []string{"80", "443"},
			},
		},
		Direction: proto.String(computepb.Firewall_INGRESS.String()),
		Name:      proto.String(firewallRuleName),
		TargetTags: []string{
			"web",
		},
		Network:     proto.String(fmt.Sprintf("projects/%s/global/networks/default", projectID)),
		Description: proto.String("Allowing TCP traffic on port 80 and 443 from Internet."),
	}

	// Insert the firewall rule
		firewallOp, err := firewallClient.Insert(ctx, &computepb.InsertFirewallRequest{
			Project:          projectID,
			FirewallResource: firewallRule,
		})
		if err != nil {
			return fmt.Errorf("unable to create firewall rule: %w", err)
		}

		// Wait for the firewall rule operation to complete
		if err := firewallOp.Wait(ctx); err != nil {
			return fmt.Errorf("unable to wait for the firewall rule operation: %w", err)
		}

		// Define and create the instance
		instanceReq := &computepb.InsertInstanceRequest{
			Project: projectID,
			Zone:    zone,
			InstanceResource: &computepb.Instance{
				Name: proto.String(instanceName),
				Disks: []*computepb.AttachedDisk{
					{
						InitializeParams: &computepb.AttachedDiskInitializeParams{
							DiskSizeGb:  proto.Int64(10),
							SourceImage: proto.String(sourceImage),
						},
						AutoDelete: proto.Bool(true),
						Boot:       proto.Bool(true),
						Type:       proto.String(computepb.AttachedDisk_PERSISTENT.String()),
					},
				},
				MachineType: proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", zone, machineType)),
				NetworkInterfaces: []*computepb.NetworkInterface{
					{
						Name: proto.String(networkName),
					},
				},
			},
		}

		// Insert the instance
		op, err := computeClient.Insert(ctx, instanceReq)
		if err != nil {
			return fmt.Errorf("unable to create instance: %w", err)
		}

		// Wait for the instance operation to complete
		if err = op.Wait(ctx); err != nil {
			return fmt.Errorf("unable to wait for the operation: %w", err)
		}

		fmt.Fprintf(w, "Instance created\n")

		return nil
	}
