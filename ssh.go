package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// createTunnel(): blah blah blah
func createTunnel(localPort, remotePort int) {
	// Configure acceptable hosts with known hosts
	hostKeyCallback, err := knownhosts.New(c.KnownHostsPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create SSH client config
	config := &ssh.ClientConfig{
		User: c.AgentID,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(c.SSHPrivKey),
		},
		HostKeyCallback: hostKeyCallback,
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.BastionIP, c.BastionPort), config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer conn.Close()

	// Create remote tunnel on all interfaces
	listener, err := conn.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", remotePort))
	if err != nil {
		log.Fatal("unable to register tcp forward: ", err)
	}
	defer listener.Close()

	for {
		remote, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			local, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
			if err != nil {
				log.Fatal(err)
			}
			defer local.Close()
			defer remote.Close()
			done := make(chan struct{}, 2)

			go func() {
				io.Copy(local, remote)
				done <- struct{}{}
			}()

			go func() {
				io.Copy(remote, local)
				done <- struct{}{}
			}()

			<-done
		}()
	}
}
