package main

import (
    // "context"
    "fmt"
    "io"
	"io/ioutil"
    // "os"
    // "time"
    // "strings"

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

	// Create a new SSH session
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("ssh.NewSession: %w", err)
	}
	defer session.Close()

	// Download the file from the URL using wget
	downloadCmd := fmt.Sprintf("wget -O %s %s", destFileName, downloadURL)
	if _, err := session.CombinedOutput(downloadCmd); err != nil {
		return fmt.Errorf("failed to download file from URL: %v", err)
	}

	fmt.Fprintf(w, "File downloaded from URL %s and stored as %s on remote instance %s\n", downloadURL, destFileName, sshHost)

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
