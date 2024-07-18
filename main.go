package main

import (
	//"bytes"
	"fmt"
	"log"
	"net/http"
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
	}
	conn, err := ssh.Dial("tcp", "192.168.1.10:22", config) // Mike's Plex server
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer conn.Close()

	// Request the remote side to open port 8080 on all interfaces.
	l, err := conn.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatal("unable to register tcp forward: ", err)
	}
	defer l.Close()

	// Serve HTTP with your SSH server acting as a reverse proxy.
	http.Serve(l, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(resp, "Hello world!\n")
	}))

}
