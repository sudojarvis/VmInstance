package main

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

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