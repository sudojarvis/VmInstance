package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os/exec"

	functions "cloud.google.com/go/functions/apiv1"
	"cloud.google.com/go/functions/apiv1/functionspb"
	"google.golang.org/api/compute/v1"
)

func randomInstanceName() string {
	return "instance-" + fmt.Sprint(rand.Intn(1000))
}





func main() {

	projectID := "cloudsec-390404"
	// zone := clea"us-central1-a"
	// instanceName := "trial-instance-1"
	// machineType := "n1-standard-1"
	// sourceImage := "projects/debian-cloud/global/images/family/debian-10"
	// networkName := "global/networks/default"
	// firewallRuleName := "allow-ssh-dummy"

	vmInstance := "trial-instance-1"

	
	Location := "us-east4" // this for cloud functions
	zone := "us-central1-a" // this for compute engine
	user := "kumar_236" //  this for the compute engine

	functionName := "test-function-1" //hardcoded for now

	ctx := context.Background()

	service, err := compute.NewService(ctx)
	if err != nil {
		log.Fatalf("Failed to create Compute Engine service: %v", err)
	}

	instance, err := service.Instances.Get(projectID, zone, vmInstance).Do()
	if err != nil {
		log.Fatalf("Failed to create Compute Engine service: %v", err)
	}

	fmt.Println("Instance details:", instance)

	fmt.Printf("Instance details:\n")
	fmt.Printf("Name: %s\n", instance.Name)
	fmt.Printf("Machine Type: %s\n", instance.MachineType)
	fmt.Printf("Status: %s\n", instance.Status)
	fmt.Printf("Internal IP: %s\n", instance.NetworkInterfaces[0].NetworkIP)
	fmt.Printf("External IP: %s\n", instance.NetworkInterfaces[0].AccessConfigs[0].NatIP)


	//take input from user
	fmt.Println("Enter the Env gen(1/2):")
	var env string
	fmt.Scanln (&env)
	// fmt.Println("Enter the function name:")

	if env == "gen2" {
		copyCloudfunctionGen2(functionName, Location)
		tranferSourceCode(functionName + ".zip", vmInstance, user, zone)

	} else {

	
		client, err := functions.NewCloudFunctionsClient(ctx)
		if err != nil {
			fmt.Printf("Failed to create client: %v", err)
			return
		}
		defer client.Close()

		// function := "projects/cloudsec-390404/locations/us-central1/functions/function-3"
		function_path := fmt.Sprintf("projects/%s/locations/%s/functions/%s", projectID, Location, functionName)


		cloudFunction, err := getCloudFunction(ctx, client, function_path)
		if err != nil {
			fmt.Println("Error getting function:", err)
			return
		}

		fmt.Println("Cloud FunctioTriggern details:", cloudFunction)


		err = downloadFunction(cloudFunction.DownloadUrl, functionName)
		if err != nil {
			fmt.Println("Error downloading function:", err)
			return
		}

		println("Downloaded function:", functionName + ".zip")


		err = tranferSourceCode(functionName + ".zip", vmInstance, user, zone)
		if err != nil {
			fmt.Println("Error transferring source code:", err)
			return
		}

	}
	
}


func getCloudFunction(ctx context.Context, client *functions.CloudFunctionsClient, functionpath string) (*functionspb.GenerateDownloadUrlResponse, error) {

	req := &functionspb.GetFunctionRequest{
		Name: functionpath,
	}

	downloadReq := &functionspb.GenerateDownloadUrlRequest{
		Name: req.Name,
	}

	cloudFunction, err := client.GenerateDownloadUrl(ctx, downloadReq)
	if err != nil {
		return nil, err
	}

	return cloudFunction, nil

}



func downloadFunction(downloadURL string, functionName string) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	functionName = functionName + ".zip"

	err = ioutil.WriteFile(functionName, body, 0644)
	if err != nil {
		return err
	}

	return nil
}




func tranferSourceCode(localFilePath, vmInstance string, user string, zone string) error {

	cmd := exec.Command("gcloud", "compute", "scp", localFilePath, user+"@"+vmInstance+":~/", "--zone", zone)
    _, err := cmd.Output()

	if err != nil {
		return err
	}

	println("Source code transferred to VM instance:", vmInstance)
	return nil
}