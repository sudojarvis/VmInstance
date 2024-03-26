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

		
		err = createInstance(os.Stdout, projectID, zone, vmInstance, machineType, sourceImage, networkName, credentialsBytes)

		if err != nil {

			removeSSHKey(privatePathKey)
			log.Fatalf("Failed to create instance: %v", err)

			c.Send("Failed to create instance")
			return
		}



		err = addPublicKeytoInstance(os.Stdout, projectID, zone, vmInstance, publicKey, user, credentialsBytes)
		if err != nil {
			log.Fatalf("Failed to add public key to instance: %v", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			c.Send("Failed to add public key to instance")
			return
		}


		ctx := context.Background()

		service, err := compute.NewService(ctx, option.WithCredentialsJSON(credentialsBytes))
		if err != nil {
			
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			log.Fatalf("Failed to create Compute Engine service: %v", err)
			c.Send("Failed to create Compute Engine service")
			return
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
			c.Send("Failed to create function client")
			return
		}
		defer client.Close()


		function_path := fmt.Sprintf("projects/%s/locations/%s/functions/%s", projectID, Location, functionName)


		cloudFunction, err := getCloudFunction(ctx, client, function_path)
		if err != nil {
			fmt.Println("Error getting function:", err)
			
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			c.Send("Failed to get cloud function")
			return
		}

		//need to remove this.
		time.Sleep(time.Second * 30) // need wait for the instance to be ready to accept the ssh connection
	
		err = downloadAndUnzipFileOnInstance(os.Stdout, cloudFunction.DownloadUrl, functionName + ".zip", external_ip, user, privatePathKey)

		if err != nil {
			fmt.Println("Error downloading and unzipping file:", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			c.Send("Failed to download and unzip file on instance")
			return
		}



		err = runGrypeOnScannerDirectory(external_ip, user, privatePathKey)

		if err != nil {
			fmt.Println("Error running Grype:", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			c.Send("Failed to run Grype on instance")
			return
		}


		err = runSemGrepOnScannerDirectory(external_ip, user, privatePathKey)

		if err != nil {
			fmt.Println("Error running SemGrep:", err)
			deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
			c.Send("Failed to run SemGrep on instance")
			return
		}



		err = deleteResources(projectID, zone, vmInstance, credentialsBytes, privatePathKey)
		if err != nil {
			fmt.Println("Error deleting resources:", err)
			c.Send("Failed to delete resources")
			return
		}

		println("Scanning completed successfully and resources deleted")

		c.Send("Scanning completed successfully and resources deleted")

	})


	err := app.Listen(":3000")
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
	println("Server running on port 3000 ...")

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


func deleteResources(projectID, zone, vmInstance string, credentialsBytes []byte, privatePathKey string) error {
	

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

	if _, err := os.Stat(privatePathKey); err == nil {
		removeSSHKey(privatePathKey)
	}

	return nil
	
}