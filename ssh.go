package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

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
			log.Printf("createTunnel(): can't create tunnel listener: %v\n", err)
		}
		go func() {
			var wg sync.WaitGroup
			wg.Add(2)
			local, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
			if err != nil {
				log.Fatal(err)
			}
			defer local.Close()
			defer remote.Close()

			go func() {
				io.Copy(local, remote)
				wg.Done()
			}()

			go func() {
				io.Copy(remote, local)
				wg.Done()
			}()

			wg.Wait()
		}()
	}
}
