package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"log"

	computeapi "google.golang.org/api/compute/v1"
	compute "cloud.google.com/go/compute/apiv1"
	"google.golang.org/api/option"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)


func deleteResources(projectID, zone, vmInstance string, credentialsBytes []byte, privatePathKey string) error {
	

	ctx := context.Background()


	service, err := computeapi.NewService(ctx, option.WithCredentialsJSON(credentialsBytes))
	if err != nil {
		log.Fatalf("Failed to create Compute Engine service: %v", err)
	}

	instance, _ := service.Instances.Get(projectID, zone, vmInstance).Do()

	if instance != nil {
		deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
		// println("Instance deleted")
	}

	if _, err := os.Stat(privatePathKey); err == nil {
		removeSSHKey(privatePathKey)
	}

	return nil
	
}

// deleteInstance sends a delete request to the Compute Engine API and waits for it to complete.
func deleteInstance(w io.Writer, projectID, zone, instanceName string, credentials []byte) error {
	// projectID := "your_project_id"
	// zone := "europe-central2-b"
	// instanceName := "your_instance_name"
	ctx := context.Background()
	instancesClient, err := compute.NewInstancesRESTClient(ctx, option.WithCredentialsJSON(credentials))
	if err != nil {
			return fmt.Errorf("NewInstancesRESTClient: %w", err)
	}
	defer instancesClient.Close()

	req := &computepb.DeleteInstanceRequest{
			Project:  projectID,
			Zone:     zone,
			Instance: instanceName,
	}

	op, err := instancesClient.Delete(ctx, req)
	if err != nil {
			return fmt.Errorf("unable to delete instance: %w", err)
	}

	if err = op.Wait(ctx); err != nil {
			return fmt.Errorf("unable to wait for the operation: %w", err)
	}

	fmt.Fprintf(w, "Instance deleted\n")

	return nil
}

func removeSSHKey(name_of_ssh_key string) error {
	
	os.Remove(name_of_ssh_key)
	println("SSH key removed:", name_of_ssh_key)
	os.Remove(name_of_ssh_key + ".pub")
	println("SSH public key removed:", name_of_ssh_key + ".pub")
	
	return nil
}