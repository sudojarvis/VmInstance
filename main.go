package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	// functions "cloud.google.com/go/functions/apiv1"

	// "cloud.google.com/go/functions/apiv1/functionspb"

	functions "cloud.google.com/go/functions/apiv2"
	"cloud.google.com/go/functions/apiv2/functionspb"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// requestBody := map[string]interface{}{
// 	"projectID": "cloudsec-390404",
// 	"vmInstance": "testvm-4",
// 	"Location": "us-east4",
// 	"zone": "us-east4-c",
// 	"user": "ashish",
// 	"path_to_json": "cloudsec-390404-be836ea29934.json",
// 	"functionName": "function-1",
// 	"Location": "us-central1",
// 	"machineType": "n1-standard-1",
// 	"sourceImage": "projects/debian-cloud/global/images/family/debian-10",
// 	"networkName": "global/networks/default",
// 	"fireWallName": "allow-ssh-2",
// }

func main() {



	projectID := "cloudsec-390404"
	vmInstance := "testvm-10"
	Location := "us-east4" // this for cloud functions

	// zone := "us-central1-a" // this for compute engine
	zone := "us-east4-c" // this for compute engine
	user := "ashish" //  this for the compute engine

	path_to_json := "cloudsec-390404-736043970d07.json"
	// functionName := "test-function-1" //hardcoded for now gen 2 function
	functionName := "function-1" // gen  function
	Location = "us-central1" // gen function




	// machineType := "n1-standard-1"
	// sourceImage := "projects/debian-cloud/global/images/family/debian-10"
	// networkName := "global/networks/default"
	fireWallName := "allow-ssh-2"

	privatePathKey := "gcp_rsa"

		// _ , publicKey, err := generateSSHKeyPair(user ,privatePathKey)
		// if err != nil {
		// 	log.Fatalf("Failed to generate SSH key pair: %v", err)
		// }

		
		// err = createInstanceWithFirewall(os.Stdout, projectID, zone, vmInstance, machineType, sourceImage, networkName, path_to_json, fireWallName)

		// if err != nil {

		// 	removeSSHKey(privatePathKey)
		// 	log.Fatalf("Failed to create instance: %v", err)
		// }


		

		// err = addPublicKeytoInstance(os.Stdout, projectID, zone, vmInstance, string(publicKey), user,path_to_json)

		// if err != nil {
		// 	log.Fatalf("Failed to add public key to instance: %v", err)
		// 	// delete the intance
		// 	deleteInstance(os.Stdout, projectID, zone, vmInstance, path_to_json)
		// 	deleteFirewallRule(os.Stdout, projectID, fireWallName, path_to_json)
		// 	removeSSHKey(privatePathKey)
		// 	return
		// }


		ctx := context.Background()

		service, err := compute.NewService(ctx, option.WithCredentialsFile(path_to_json))
		if err != nil {
			
			deleteInstance(os.Stdout, projectID, zone, vmInstance, path_to_json)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, path_to_json)
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



		client, err := functions.NewFunctionClient(ctx, option.WithCredentialsFile(path_to_json))
		if err != nil {
			fmt.Printf("Failed to create client: %v", err)
			deleteInstance(os.Stdout, projectID, zone, vmInstance, path_to_json)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, path_to_json)
			removeSSHKey(privatePathKey)
			// return
		}
		defer client.Close()


		function_path := fmt.Sprintf("projects/%s/locations/%s/functions/%s", projectID, Location, functionName)


		cloudFunction, err := getCloudFunction(ctx, client, function_path)
		if err != nil {
			fmt.Println("Error getting function:", err)
			// delete the intance
			deleteInstance(os.Stdout, projectID, zone, vmInstance, path_to_json)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, path_to_json)
			removeSSHKey(privatePathKey)
			// return
		}


		fmt.Println("Cloud Function details:", cloudFunction)
		err = downloadFunction(cloudFunction.DownloadUrl, functionName)


		if err != nil {
			deleteInstance(os.Stdout, projectID, zone, vmInstance, path_to_json)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, path_to_json)
			removeSSHKey(privatePathKey)
			
			fmt.Println("Error downloading function:", err)
			// return
		}
		

		err = copySourceCode(functionName + ".zip", external_ip, user, zone, privatePathKey)
		if err != nil {
			fmt.Println("Error copying source code:", err)
			// deleteInstance(os.Stdout, projectID, zone, vmInstance, path_to_json)
			// deleteFirewallRule(os.Stdout, projectID, fireWallName, path_to_json)
			// removeSSHKey(privatePathKey)
			return
		}


		// os.Remove(functionName + ".zip")
		// println("Source code copied:", functionName + ".zip")
		// func downloadAndUnzipFileOnInstance(w io.Writer, downloadURL, destFileName, sshHost, sshUser, privateKeyPath string) error {


		// downloadAndUnzipFileOnInstance(os.Stdout, cloudFunction.DownloadUrl, functionName + ".zip", external_ip, user, privatePathKey)

		// time.Sleep(time.Second * 30)


		err = tranferSourceCode(functionName, vmInstance, user, zone, privatePathKey, external_ip, cloudFunction.DownloadUrl)

		if err != nil {
			fmt.Println("Error transferring source code:", err)
			deleteInstance(os.Stdout, projectID, zone, vmInstance, path_to_json)
			deleteFirewallRule(os.Stdout, projectID, fireWallName, path_to_json)
			removeSSHKey(privatePathKey)

			return
		}

	// })


	// app.Listen(":3000")



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


// func tranferSourceCode(localFilePath, vmInstance string, user string, zone string) error {

// 	cmd := exec.Command("gcloud", "compute", "scp", localFilePath, user+"@"+vmInstance+":~/", "--zone", zone)
//     _, err := cmd.Output()

// 	if err != nil {
// 		return err
// 	}

// 	println("Source code transferred to VM instance:", vmInstance)
// 	return nil
// }

func copySourceCode(functionName, external_ip, user, zone string, privateKeyPath string) error {
	
	scpCmd := fmt.Sprintf("scp -i %s -o StrictHostKeyChecking=no %s %s@%s:~/", privateKeyPath, functionName, user, external_ip)
    cmd := exec.Command("bash", "-c", scpCmd)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error copying file to VM instance: %v", err)
    }

	return nil
}


func tranferSourceCode(functionName, vmInstance, user, zone, privateKeyPath string, external_ip string, downloadURL string) error {
	// Create Compute Engine service\\\
	// computeService, err := compute.NewService(context.Background())
	// if err != nil {
	// 	return fmt.Errorf("error creating compute service: %v", err)
	// }

	// Get instance info
	// instanceInfo, err := computeService.Instances.Get("cloudsec-390404", zone, vmInstance).Do()
	// if err != nil {
	// 	return fmt.Errorf("error getting instance info: %v", err)
	// }

	key, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("error reading private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("error parsing private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	client, err := ssh.Dial("tcp", external_ip+":22", config)
	if err != nil {
		//delete the instance
		//delete the firewall
		//delete the ssh key
		//delete the function

		return fmt.Errorf("error dialing: %v", err)
	}
	defer client.Close()


	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating session: %v", err)
	}
	defer session.Close()

	// if err := session.Run("curl -o " + functionName + ".zip " + downloadURL); err != nil {
	// 	return fmt.Errorf("error downloading function: %v", err)
	// }

	// defer session.Close()

	// if err := session.Run()
	

	// install unzip
	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating session: %v", err)
	}

	if err := session.Run("sudo apt-get install unzip"); err != nil {
		return fmt.Errorf("error installing unzip: %v", err)
	}

	defer session.Close()

	// session, err = client.NewSession()
	// if err != nil {
	// 	return fmt.Errorf("error creating session: %v", err)
	// }

	// if err := session.Run("sudo apt-get install p7zip-full"); err != nil {
	// 	return fmt.Errorf("error installing 7zip: %v", err)
	// }


	//create directory and install grype
	session, err = client.NewSession()
    if err != nil {
        log.Fatalf("Failed to create session: %v", err)
    }
    defer session.Close()

    if err := session.Run("mkdir -p ~/scanner && sudo curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sudo sh -s -- -b /usr/local/bin"); err != nil {
        log.Fatalf("Error creating directory or installing grype: %v", err)
    }

	println("Directory created and grype installed")


	//unzip the file in the scanner directory

	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating session: %v", err)
	}

	defer session.Close()

	println("Unzipping file", functionName + ".zip")

	if err := session.Run("sudo unzip -o " + functionName + ".zip -d ~/scanner"); err != nil {
		return fmt.Errorf("error unzipping file: %v", err)
	} 
	println("File unzipped successfully")


	//execute the grype command

	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating session: %v", err)
	}

	defer session.Close()

	if err := session.Run("grype -v ~/scanner --output json > output.json"); err != nil {
		return fmt.Errorf("error executing grype command: %v", err)
	}

	println("Grype command executed successfully")

	// send the output.json to backend api

	// session, err = client.NewSession()
	// if err != nil {
	// 	return fmt.Errorf("error creating session: %v", err)
	// }

	// if err := session.Run("curl -X POST -H \"Content-Type: application/json\" -d @output.json http://localhost:8080/api/v1/scan"); err != nil {

	// 	return fmt.Errorf("error sending output.json to backend api: %v", err)
	// }
	
    // fmt.Println("Commands executed successfully")


	return nil

}


func removeSSHKey(name_of_ssh_key string) error {
	
	os.Remove(name_of_ssh_key)
	println("SSH key removed:", name_of_ssh_key)
	os.Remove(name_of_ssh_key + ".pub")
	println("SSH public key removed:", name_of_ssh_key + ".pub")
	
	return nil
}
