package main

import (
	"VmInstance/cf"
	"VmInstance/sshFunctions"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	// functions "cloud.google.com/go/functions/apiv1"

	// "cloud.google.com/go/functions/apiv1/functionspb"

	functions "cloud.google.com/go/functions/apiv2"

	"github.com/gofiber/fiber/v2"
	compute "google.golang.org/api/compute/v1"
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
	privatePathKey := "gcp_rsa"

	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		// Get the body of the request
		body := []byte(c.Body())

		// Parse the body as JSON
		var requestBody map[string]interface{}
		err := json.Unmarshal(body, &requestBody)
		if err != nil {
			return c.Status(400).Send([]byte("Invalid request body"))
		}

		credentials, ok := requestBody["credentials"].(map[string]interface{})
		if !ok {
			return c.Status(400).Send([]byte("Missing or invalid 'credentials' key in request body"))
		}

		println("Credentials:", credentials)

		credentialsJSON, err := json.Marshal(credentials)
		if err != nil {
			return c.Status(500).Send([]byte("Failed to marshal credentials to JSON"))
		}

		credentialsBytes = credentialsJSON

		projectID = credentials["project_id"].(string)

		Location, ok := requestBody["Location"].(string)
		if !ok {
			return c.Status(400).Send([]byte("Missing 'Location' key in request body"))
		}

		functionName, ok := requestBody["functionName"].(string)
		if !ok {
			return c.Status(400).Send([]byte("Missing 'functionName' key in request body"))
		}

		user, ok = requestBody["user"].(string)
		if !ok {
			return c.Status(400).Send([]byte("Missing 'user' key in request body"))
		}

		zone, ok = requestBody["zone"].(string)
		if !ok {
			return c.Status(400).Send([]byte("Missing 'zone' key in request body"))
		}

		_, publicKey, err := generateSSHKeyPair(user, privatePathKey)
		if err != nil {
			log.Fatalf("Failed to generate SSH key pair: %v", err)
			return err
		}

		err = createInstance(os.Stdout, projectID, zone, vmInstance, machineType, sourceImage, networkName, credentialsBytes)
		if err != nil {
			removeSSHKey(privatePathKey)
			log.Fatalf("Failed to create instance: %v", err)
			return c.Send([]byte("Failed to create instance"))
		}

		err = addPublicKeytoInstance(os.Stdout, projectID, zone, vmInstance, publicKey, user, credentialsBytes)
		if err != nil {
			log.Fatalf("Failed to add public key to instance: %v", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			return c.Send([]byte("Failed to add public key to instance"))
		}

		ctx := context.Background()

		service, err := compute.NewService(ctx, option.WithCredentialsJSON(credentialsBytes))
		if err != nil {
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			log.Fatalf("Failed to create Compute Engine service: %v", err)
			return c.Send([]byte("Failed to create Compute Engine service"))
		}

		instance, err := service.Instances.Get(projectID, zone, vmInstance).Do()
		if err != nil {
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			log.Fatalf("Failed to create Compute Engine service: %v", err)
		}

		fmt.Println("Instance details:", instance)

		external_ip := instance.NetworkInterfaces[0].AccessConfigs[0].NatIP

		println("External IP:", external_ip)

		client, err := functions.NewFunctionClient(ctx, option.WithCredentialsJSON(credentialsBytes))
		if err != nil {
			fmt.Printf("Failed to create client: %v", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			return c.Send([]byte("Failed to create function client"))
		}
		defer client.Close()

		function_path := fmt.Sprintf("projects/%s/locations/%s/functions/%s", projectID, Location, functionName)

		cloudFunction, err := cf.GetCloudFunction(ctx, client, function_path)
		if err != nil {
			fmt.Println("Error getting function:", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			return c.Send([]byte("Failed to get cloud function"))
		}

		//need to remove this.
		time.Sleep(time.Second * 30) // need wait for the instance to be ready to accept the ssh connection

		err = sshFunctions.DownloadAndUnzipFileOnInstance(os.Stdout, cloudFunction.DownloadUrl, functionName+".zip", external_ip, user, privatePathKey)
		if err != nil {
			fmt.Println("Error downloading and unzipping file:", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			return c.Send([]byte("Failed to download and unzip file on instance"))
		}

		err = runGrypeOnScannerDirectory(external_ip, user, privatePathKey)
		if err != nil {
			fmt.Println("Error running Grype:", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			return c.Send([]byte("Failed to run Grype on instance"))
		}

		err = runSemGrepOnScannerDirectory(external_ip, user, privatePathKey)
		if err != nil {
			fmt.Println("Error running SemGrep:", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			return c.Send([]byte("Failed to run SemGrep on instance"))
		}

		err = deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
		if err != nil {
			fmt.Println("Error deleting resources:", err)
			return c.Send([]byte("Failed to delete resources"))
		}

		println("Scanning completed successfully and resources deleted")

		return c.Send([]byte("Scanning completed successfully and resources deleted"))
	})

	err := app.Listen(":3000")
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
	println("Server running on port 3000 ...")

}