package main

import (
	"fmt"
	"os/exec"
	"regexp"

	"gopkg.in/yaml.v2"
)

func copyCloudfunctionGen2(functionName string, Location string) {

   
    

    println("-------------------------------------------")

	command := "gcloud"
    args := []string{"functions", "describe", fmt.Sprintf("%s", functionName), "--region", Location}


    out, err := exec.Command(command, args...).Output()
	if err != nil {
		fmt.Println("Error executing command:", err)
		fmt.Println("Output:", string(out))
		return
	}

	bucket := ""
	object := ""
	// // generation := ""

	sourcePattern := regexp.MustCompile(`source:\s+storageSource:\s+bucket:\s+(?P<bucket>\S+)\s+generation:\s+'(?P<generation>\d+)'\s+object:\s+(?P<object>\S+)`)
    matches := sourcePattern.FindStringSubmatch(string(out))



    if len(matches) > 0 {
        bucket = matches[1]
        generation := matches[2]
        object = matches[3]

        fmt.Println("Source:")
        fmt.Println("Bucket:", bucket)
        fmt.Println("Generation:", generation)
        fmt.Println("Object:", object)
    } else {
        fmt.Println("Source information not found.")
    }

	//////////////////////////////////////////////////////////////////
	command1 := "gsutil"
    args2 := []string{"cp", fmt.Sprintf("gs://%s/%s", bucket, object), fmt.Sprintf("%s.zip", functionName)}

    // Execute the command
    cmd := exec.Command(command1, args2...)
    out1, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Error executing command:", err)
        fmt.Println("Combined Output:", string(out1))
        return
    }

    fmt.Println("Output:", string(out))

}


func GFunctionDescriptionEnv(functionName string, Location string) (string) {

    command := "gcloud"
	args := []string{"functions", "describe", functionName, "--region", Location, "--format=yaml"}

	out, err := exec.Command(command, args...).Output()
	if err != nil {
		fmt.Println("Error executing command:", err)
		fmt.Println("Output:", string(out))
		return ""
	}

	var functionInfo map[string]interface{}
	if err := yaml.Unmarshal(out, &functionInfo); err != nil {
		fmt.Println("Error parsing YAML:", err)
		return ""
	}

	environment, ok := functionInfo["environment"].(string)
	if !ok {
		fmt.Println("Environment not found in YAML")
		return ""
	}
	fmt.Println("Environment:", environment)

    return environment
}


