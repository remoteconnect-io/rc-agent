package main

import (
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// TODO: move these constants to config struct
const (
	srvIP   = "192.168.1.10"
	srvPort = "22"
	userID  = "mmillsap"
)

// pre-configured parameters: user, hostIP, ssh_cert, -- all these should be in a config struct, initialized by config function at beginning of program
// passed parameters: server_port, local_port
func creatTunnel(localPort, remotePort string) {

	// TODO: move to separate function or add to config struct
	// Get SSH private key and create the Signer
	key, err := os.ReadFile("/home/mmillsap/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read SSH private key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse SSH private key: %v", err)
	}

	// TODO: move to config struct???
	// Configure acceptable hosts with known hosts
	hostKeyCallback, err := knownhosts.New("known_hosts")
	if err != nil {
		log.Fatal(err)
	}

	// Create SSH client config
	config := &ssh.ClientConfig{
		User: userID,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}
	conn, err := ssh.Dial("tcp", srvIP+":"+srvPort, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer conn.Close()

	// Create remote tunnel on all interfaces
	listener, err := conn.Listen("tcp", "0.0.0.0:"+remotePort)
	if err != nil {
		log.Fatal("unable to register tcp forward: ", err)
	}
	defer listener.Close()

	// Copy I/O from tunnel to local port
	for {
		remote, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			local, err := net.Dial("tcp", "localhost:"+localPort)
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
