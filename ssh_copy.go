package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/sftp"
)

// downloadAndUnzipFileOnInstance downloads an object to a file on a remote VM instance, and then unzips it.
func downloadAndUnzipFileOnInstance(w io.Writer, downloadURL, destFileName, sshHost, sshUser, privateKeyPath string) error {
	// Connect to the VM instance via SSH
	sshClient, err := sshConnect(sshHost, sshUser, privateKeyPath)
	if err != nil {
		return fmt.Errorf("sshConnect: %w", err)
	}
	defer sshClient.Close()

	// Create a new SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("sftp.NewClient: %w", err)
	}
	defer sftpClient.Close()

	// Create a new file on the remote instance to store the downloaded file
	remoteFile, err := sftpClient.Create(destFileName)
	if err != nil {
		return fmt.Errorf("sftpClient.Create: %w", err)
	}
	defer remoteFile.Close()

	// Create a directory called "scanner" on the remote instance
	err = sshRunCommand(sshClient, "mkdir -p scanner")
	if err != nil {
		return fmt.Errorf("failed to create directory 'scanner': %w", err)
	}

	// Download the file from the URL and write it to the remote file
	response, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("http.Get: %w", err)
	}
	defer response.Body.Close()

	if _, err := io.Copy(remoteFile, response.Body); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	fmt.Fprintf(w, "File downloaded from URL %s and stored as %s on remote instance %s\n", downloadURL, destFileName, sshHost)

	// Install unzip on the remote VM instance
	err = sshRunCommand(sshClient, "sudo apt-get update && sudo apt-get install -y unzip")
	if err != nil {
		return fmt.Errorf("failed to install unzip: %w", err)
	}

	// Unzip the downloaded file into the "scanner" directory
	_, err = sshRunCommandWithOutput(sshClient, fmt.Sprintf("sudo unzip -o %s -d scanner/", destFileName))
	if err != nil {
		return fmt.Errorf("failed to unzip file: %w", err)
	}
	println("File unzipped into scanner directory")

    
	return nil
}

