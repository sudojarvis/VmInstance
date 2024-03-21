package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"github.com/pkg/sftp"
	"os"
)


func runGrypeOnScannerDirectory(sshHost, sshUser, privateKeyPath string) error {

    sshClient, err := sshConnect(sshHost, sshUser, privateKeyPath)
	if err != nil {
		return fmt.Errorf("sshConnect: %w", err)
	}
	defer sshClient.Close()
    //installing the grype
    _, err = sshRunCommandWithOutput(sshClient, "sudo curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sudo sh -s -- -b /usr/local/bin")
    if err != nil {
        return fmt.Errorf("failed to install grype: %w", err)
    }
    println("Grype installed successfully")

    // run grype on the scanner directory

    _, err = sshRunCommandWithOutput(sshClient, "grype -v ~/scanner --output json > grype_output.json")
	if err != nil {
		return fmt.Errorf("failed to run grype: %w", err)
	}
	println("Grype ran successfully")

	// Downloading Grype report
	err = downloadFile(sshClient, "grype_output.json", "grype_output.json")
	if err != nil {
		return fmt.Errorf("failed to download Grype report: %w", err)
	}
	fmt.Println("Grype report downloaded successfully")

	return nil
}

func runSemGrepOnScannerDirectory(sshHost, sshUser, privateKeyPath string) error {

    sshClient, err := sshConnect(sshHost, sshUser, privateKeyPath)
	if err != nil {
		return fmt.Errorf("sshConnect: %w", err)
	}
	defer sshClient.Close()

	//installing pip and python3

	_, err = sshRunCommandWithOutput(sshClient, "sudo apt-get update && sudo apt-get install -y python3-pip")
	if err != nil {
		return fmt.Errorf("failed to install pip and python3: %w", err)
	}


    //installing the semgrep
    _, err = sshRunCommandWithOutput(sshClient, "sudo python3 -m pip install semgrep")
    if err != nil {
        return fmt.Errorf("failed to install semgrep: %w", err)
	}

	println("Semgrep installed successfully")

	// run semgrep on the scanner directory
	_, err = sshRunCommandWithOutput(sshClient, "sudo semgrep scan ~/scanner --json > semgrep_output.json")
	if err != nil {
		return fmt.Errorf("failed to run semgrep: %w", err)
	}

	println("Semgrep ran successfully")

	// Downloading Semgrep report
	err = downloadFile(sshClient, "semgrep_output.json", "semgrep_output.json")
	if err != nil {
		return fmt.Errorf("failed to download Semgrep report: %w", err)
	}
	fmt.Println("Semgrep report downloaded successfully")

	return 	nil
}

func downloadFile(client *ssh.Client, remotePath, localPath string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = srcFile.WriteTo(destFile)
	if err != nil {
		return err
	}

	return nil
}