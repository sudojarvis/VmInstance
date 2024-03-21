package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	// functions "cloud.google.com/go/functions/apiv1"

	// "cloud.google.com/go/functions/apiv1/functionspb"

	functions "cloud.google.com/go/functions/apiv2"
	"cloud.google.com/go/functions/apiv2/functionspb"
	"github.com/gofiber/fiber"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)


var (
	projectID string
	Location string
	// functionName string
	user string
	zone string
	credentialsBytes []byte
)



func randomInstanceName() string {
	rand := time.Now().UnixNano()
	return fmt.Sprintf("instance-%d", rand)
}





func main() {



	// projectID := "cloudsec-390404"
	vmInstance :=randomInstanceName()
	println("Instance name:", vmInstance)
	// Location := "us-east4" // this for cloud functions

	// // zone := "us-central1-a" // this for compute engine
	// zone := "us-east4-c" // this for compute engine
	// user := "ashish" //  this for the compute engine

	// path_to_json := "cloudsec-390404-736043970d07.json"
	// functionName := "test-function-1" //hardcoded for now gen 2 function
	// functionName := "function-1" // gen  function
	// Location = "us-central1" // gen function



	//haedcoded for values 
	machineType := "n1-standard-1"
	sourceImage := "projects/debian-cloud/global/images/family/debian-10"
	networkName := "global/networks/default"
	fireWallName := "allow-ssh-2"
	privatePathKey := "gcp_rsa"



	app := fiber.New()


	app.Post("/test", func(c *fiber.Ctx) {
		// Get the body of the request
		body := []byte(c.Body())

		// Parse the body as JSON
		var requestBody map[string]interface{}
		err := json.Unmarshal(body, &requestBody)
		if err != nil {
			c.Status(400).Send("Invalid request body")
			return
		}

		credentials, ok := requestBody["credentials"].(map[string]interface{})
		if !ok {
			c.Status(400).Send("Missing or invalid 'credentials' key in request body")
			return
		}

		println("Credentials:", credentials)

		credentialsJSON, err := json.Marshal(credentials)
		if err != nil {
			c.Status(500).Send("Failed to marshal credentials to JSON")
			return
		}


		credentialsBytes = credentialsJSON
	
		projectID = credentials["project_id"].(string)

		Location, ok := requestBody["Location"].(string)
		if !ok {
			c.Status(400).Send("Missing 'Location' key in request body")
			return
		}
	
		functionName, ok := requestBody["functionName"].(string)
		if !ok {
			c.Status(400).Send("Missing 'functionName' key in request body")
			return
		}

		user, ok = requestBody["user"].(string)
		if !ok {
			c.Status(400).Send("Missing 'user' key in request body")
			return
		}
	
		zone, ok = requestBody["zone"].(string)
		if !ok {
			c.Status(400).Send("Missing 'zone' key in request body")
			return
		}


		_ , publicKey, err := generateSSHKeyPair(user ,privatePathKey)
		if err != nil {
			log.Fatalf("Failed to generate SSH key pair: %v", err)
			return
		}

		
		err = createInstanceWithFirewall(os.Stdout, projectID, zone, vmInstance, machineType, sourceImage, networkName, credentialsBytes, fireWallName)

		if err != nil {

			removeSSHKey(privatePathKey)
			log.Fatalf("Failed to create instance: %v", err)
			return
		}



		err = addPublicKeytoInstance(os.Stdout, projectID, zone, vmInstance, publicKey, user, credentialsBytes)
		if err != nil {
			log.Fatalf("Failed to add public key to instance: %v", err)
			// delete the intance
			deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, credentialsBytes)
			removeSSHKey(privatePathKey)
			return
		}


		ctx := context.Background()

		service, err := compute.NewService(ctx, option.WithCredentialsJSON(credentialsBytes))
		if err != nil {
			
			deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, credentialsBytes)
			removeSSHKey(privatePathKey)
			log.Fatalf("Failed to create Compute Engine service: %v", err)
		}

		instance, err := service.Instances.Get(projectID, zone, vmInstance).Do()
		if err != nil {
			log.Fatalf("Failed to create Compute Engine service: %v", err)
		}

		fmt.Println("Instance details:", instance)

		external_ip := instance.NetworkInterfaces[0].AccessConfigs[0].NatIP

		println("External IP:", external_ip)



		client, err := functions.NewFunctionClient(ctx, option.WithCredentialsJSON(credentialsBytes))
		if err != nil {
			fmt.Printf("Failed to create client: %v", err)
			deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, credentialsBytes)
			removeSSHKey(privatePathKey)
			return
		}
		defer client.Close()


		function_path := fmt.Sprintf("projects/%s/locations/%s/functions/%s", projectID, Location, functionName)


		cloudFunction, err := getCloudFunction(ctx, client, function_path)
		if err != nil {
			fmt.Println("Error getting function:", err)
			// delete the intance
			deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, credentialsBytes)
			removeSSHKey(privatePathKey)
			return
		}

		//need to remove this.
		time.Sleep(time.Second * 30) // need wait for the instance to be ready to accept the ssh connection
	
		err = downloadAndUnzipFileOnInstance(os.Stdout, cloudFunction.DownloadUrl, functionName + ".zip", external_ip, user, privatePathKey)

		if err != nil {
			fmt.Println("Error downloading and unzipping file:", err)
			deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, credentialsBytes)
			removeSSHKey(privatePathKey)
			return
		}



		err = runGrypeOnScannerDirectory(external_ip, user, privatePathKey)

		if err != nil {
			fmt.Println("Error running Grype:", err)
			deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, credentialsBytes)
			removeSSHKey(privatePathKey)
			return
		}


		err = runSemGrepOnScannerDirectory(external_ip, user, privatePathKey)

		if err != nil {
			fmt.Println("Error running SemGrep:", err)
			deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, credentialsBytes)
			removeSSHKey(privatePathKey)
			return
		}



	})


	app.Get("/delete", func(c *fiber.Ctx) {

		// println("credentialsBytes:", credentialsBytes)
		// print("projectID:", projectID)
		// print("zone:", zone)
		// print("vmInstance:", vmInstance)
		// print("fireWallName:", fireWallName)

		//
		ctx := context.Background()
		

		service, err := compute.NewService(ctx, option.WithCredentialsJSON(credentialsBytes))
		if err != nil {
			log.Fatalf("Failed to create Compute Engine service: %v", err)
		}

		instance, _ := service.Instances.Get(projectID, zone, vmInstance).Do()

		if instance != nil {
			deleteInstance(os.Stdout, projectID, zone, vmInstance, credentialsBytes)
			// println("Instance deleted")
		}

		firewall, _ := service.Firewalls.Get(projectID, fireWallName).Do()
		
		if firewall != nil {
			deleteFirewallRule(os.Stdout, projectID, fireWallName, credentialsBytes)
			// println("Firewall rule deleted")
		}


		if _, err := os.Stat(privatePathKey); err == nil {
			removeSSHKey(privatePathKey)

			println("SSH key removed")
		}

		c.Send("Instance deleted , firewall rule deleted and SSH key removed")
	})



	println("Server running on port 3000 ...")
	app.Listen(":3000")

}


func getCloudFunction(ctx context.Context, client *functions.FunctionClient, functionpath string) (*functionspb.GenerateDownloadUrlResponse, error) {

	req := &functionspb.GetFunctionRequest{
		Name: functionpath,
	}

	downloadReq := &functionspb.GenerateDownloadUrlRequest{
		Name: req.Name,
	}

	
	info, err := client.GetFunction(ctx, req)
	if err != nil {
		return nil, err
	}

	fmt.Println("Cloud Function details:", info)

	cloudFunction, err := client.GenerateDownloadUrl(ctx, downloadReq)
	if err != nil {
		return nil, err
	}
	return cloudFunction, nil

}