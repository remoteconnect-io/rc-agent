package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func main() {
	// var hostKey ssh.PublicKey
	key, err := os.ReadFile("/home/mmillsap/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read SSH private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse SSH private key: %v", err)
	}

	// Configure known hosts
	hostKeyCallback, err := knownhosts.New("known_hosts")
	if err != nil {
		log.Fatal(err)
	}

	// Create SSH client config using SSH keys
	config := &ssh.ClientConfig{
		User: "mmillsap",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		// HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", "192.168.1.10:22", config) // Mike's Plex server
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("/usr/bin/hostname"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Printf("Remotely connected to: %s", b.String())
}
