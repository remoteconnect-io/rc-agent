package main

import "log"

var c Config

func init() {
	err := getConfig(&c, "./config/config.json")
	if err != nil {
		log.Fatalf("error getting configuration settings: %v", err)
	}
}

func main() {
	createTunnel(22, 8080)
}
