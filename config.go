package main

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	AgentID        string `json:"agent_id"`
	ServerIP       string `json:"server_ip"`
	ServerPort     int    `json:"server_port"`
	BastionIP      string `json:"bastion_ip"`
	BastionPort    int    `json:"bastion_port"`
	SSHPrivKey     ssh.Signer
	KnownHostsPath string `json:"known_hosts_path"`
	KeyPath        string `json:"key_path"`
	KeySize        int    `json:"key_size"`
	JwtExpMinutes  int    `json:"jwt_exp_minutes"`
	Heartbeat      int    `josn:"heartbeat"`
}

// getConfig() loads configuration parameters from the config file
func getConfig(conf *Config, cfgPath string) error {
	j, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("getConfig(): %w", err)
	}
	err = json.Unmarshal(j, conf)
	if err != nil {
		return fmt.Errorf("getConfig(): %w", err)
	}

	// TODO: During app startup, need to check if key pair exists,
	//       or create new & register
	// Get SSH private key and create the Signer
	key, err := os.ReadFile("/home/mmillsap/.ssh/id_rsa")
	if err != nil {
		return fmt.Errorf("getConfig(): unable to read SSH private key: %w", err)
	}
	conf.SSHPrivKey, err = ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("getConfig(): unable to parse SSH private key: %w", err)
	}

	return nil
}
