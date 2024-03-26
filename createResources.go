package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	compute "cloud.google.com/go/compute/apiv1"
	"golang.org/x/oauth2/google"
	computeapi "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)


// func firewallRuleExists(ctx context.Context, projectID, firewallName string) (bool, error) {
// 	firewallClient, err := compute.NewFirewallsRESTClient(ctx)
// 	if err != nil {
// 		return false, fmt.Errorf("NewFirewallsRESTClient: %w", err)
// 	}
// 	defer firewallClient.Close()

// 	req := &computepb.GetFirewallRequest{
// 		Project: projectID,
// 		Firewall: firewallName,
// 	}

// 	_, err = firewallClient.Get(ctx, req)
// 	if err != nil {
// 		if status.Code(err) == codes.NotFound {
// 			return false, nil
// 		}
// 		return false, fmt.Errorf("unable to get firewall rule: %w", err)
// 	}

// 	return true, nil
// }

// createInstance sends an instance creation request to the Compute Engine API and waits for it to complete.
func createInstance(w io.Writer, projectID, zone, instanceName, machineType, sourceImage, networkName string, credentials []byte) error {
        // projectID := "cloudsec-390404"
		// zone := "us-east4-c" // Change this to your desired zone
		// instanceName := "test-vm-inst-5"
		// machineType := "n1-standard-1" // Change this to your desired machine type
        // sourceImage := "projects/cloudsec-390404/global/images/image-1"
        // networkName := "global/networks/default"

        // firewallName := "allow-ssh-ingress-from-iap"


        tags := []string{"http-server", "https-server"}

        ctx := context.Background()
        instancesClient, err := compute.NewInstancesRESTClient(ctx, option.WithCredentialsJSON(credentials))
        if err != nil {
                return fmt.Errorf("NewInstancesRESTClient: %w", err)
        }
        defer instancesClient.Close()

        req := &computepb.InsertInstanceRequest{
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
                                        AccessConfigs: []*computepb.AccessConfig{
                                                {
                                                        Name: proto.String("External NAT"),
                                                        Type: proto.String(computepb.AccessConfig_ONE_TO_ONE_NAT.String()),
                                                },
                                        },
                                        
                                },
                        },
                        Tags: &computepb.Tags{
                                Items: tags,
                        },
                        
                },
        }

        op, err := instancesClient.Insert(ctx, req)
        if err != nil {
                return fmt.Errorf("unable to create instance: %w", err)
        }

        if err = op.Wait(ctx); err != nil {
                return fmt.Errorf("unable to wait for the operation: %w", err)
        }

        fmt.Fprintf(w, "Instance created successfully \n")

        return nil
}






func generateSSHKeyPair(username string, path string) ([]byte, []byte, error) {

	// usr, err := user.Current()
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("failed to get user's home directory: %v", err)
	// }
	// sshDir := usr.HomeDir + "/.ssh/"
        
	// privKeyPath := sshDir + "gcp_rsa"
	// pubKeyPath := sshDir + "gcp_rsa.pub"

	// Check if the key files exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Generate RSA private key
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "2048", "-f", path, "-N", "", "-C", username)
		err := cmd.Run()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
		}
	}

	
	privKey, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %v", err)
	}

	pubKey, err := ioutil.ReadFile(path + ".pub")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read public key: %v", err)
	}

	return privKey, pubKey, nil
}







func addPublicKeytoInstance(w io.Writer, projectID string, zone string, instanceName string, publicKey []byte, username string, credentials []byte) error {
    ctx := context.Background()
        //convert to string

        publicKeyString := string(publicKey)


    

    // client, err := google.DefaultClient(ctx, computeapi.CloudPlatformScope)
    // if err != nil {
    //     return fmt.Errorf("failed to create compute client: %w", err)
    // }

	// data, err := ioutil.ReadFile(path_to_json)

	// if err != nil {
	// 	return fmt.Errorf("failed to read service account key file: %w", err)
	// }

        data := credentials

	conf, err := google.JWTConfigFromJSON(data, computeapi.CloudPlatformScope)

	if err != nil {
		return fmt.Errorf("failed to create JWT config: %w", err)
	}
	
	client := conf.Client(ctx)

    computeService, err := computeapi.New(client)
    if err != nil {
        return fmt.Errorf("failed to create compute service: %w", err)
    }

 
    instanceResource := fmt.Sprintf("projects/%s/zones/%s/instances/%s", projectID, zone, instanceName)
    instanceInfo, err := computeService.Instances.Get(projectID, zone, instanceName).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("failed to get instance info: %v", err)
    }

 
    fmt.Fprintf(w, "Instance details:\n")
    fmt.Fprintf(w, "Name: %s\n", instanceInfo.Name)
    fmt.Fprintf(w, "Machine Type: %s\n", instanceInfo.MachineType)

    publicKeyString= fmt.Sprintf("%s:%s", username, publicKeyString)
    metadataKey := "ssh-keys"
    metadataItem := &computeapi.MetadataItems{
        Key:   metadataKey,
        Value: &publicKeyString,
		
    }
    instanceInfo.Metadata.Items = append(instanceInfo.Metadata.Items, metadataItem)

 
    _, err = computeService.Instances.SetMetadata(projectID, zone, instanceName, instanceInfo.Metadata).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("failed to update instance metadata: %v", err)
    }
    fmt.Fprintf(w, "Metadata updated\n")

	

    fmt.Printf("Public key added to instance metadata %s\n", instanceResource)

    return nil
}