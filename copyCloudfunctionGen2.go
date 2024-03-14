package main

import (
	"fmt"
	"os/exec"
	"regexp"
)

func copyCloudfunctionGen2(functionName string, Location string) {


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

    // Print the output
    fmt.Println("Output:", string(out))

}