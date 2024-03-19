package main

import (
    // "context"
    "fmt"
    "io"
	"io/ioutil"
    // "os"
    // "time"
    // "strings"
	// "golang.org/x/crypto/ssh/terminal"
	// "golang.org/x/crypto/ssh/agent"
	"github.com/pkg/sftp"
	"net/http"

    // "cloud.google.com/go/storage"
    "golang.org/x/crypto/ssh"
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

// sshConnect establishes an SSH connection to the specified host with the provided private key.
func sshConnect(sshHost, sshUser, privateKeyPath string) (*ssh.Client, error) {
    privateKey, err := ioutil.ReadFile(privateKeyPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read private key: %v", err)
    }

    signer, err := ssh.ParsePrivateKey(privateKey)
    if err != nil {
        return nil, fmt.Errorf("failed to parse private key: %v", err)
    }

    config := &ssh.ClientConfig{
        User: sshUser,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(signer),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    return ssh.Dial("tcp", sshHost+":22", config)
}

// sshRunCommand runs a command on the remote VM instance.
func sshRunCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Run(command)
}

// sshRunCommandWithOutput runs a command on the remote VM instance and returns its output.
func sshRunCommandWithOutput(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", err
	}

	return string(output), nil
}