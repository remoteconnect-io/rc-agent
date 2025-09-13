package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AgentID       string
	ServerIP      string `mapstructure:"server_ip"`
	ServerPort    int    `mapstructure:"server_port"`
	JwtExpMinutes int    `mapstructure:"jwt_exp_minutes"`
	Heartbeat     int    `mapstructure:"heartbeat"`
	KeyBits       int    `mapstructure:"key_bits"`
	Private       []byte
	Public        []byte
}

// LoadConfig loads configuration from JSON file or environment variables,
// with environment variables taking precedence.
func LoadConfig(configFile string) (*Config, error) {

	v := viper.New()

	// Set defaults
	v.SetDefault("key_path", "keys")
	v.SetDefault("key_bits", 4096)
	v.SetDefault("jwt_exp_minutes", 5)
	v.SetDefault("heartbeat", 5)

	// Read from file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	v.SetEnvPrefix("PROV") // environmental variable prefix
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Check if keys exist
	var noKeys bool
	var err error
	cfg.Private, err = os.ReadFile("private.pem")
	if err != nil {
		noKeys = true
	}
	cfg.Public, err = os.ReadFile("public.pem")
	if err != nil {
		noKeys = true
	}

	// TODO: I think I should not use directory for keys, just put in root path
	// If either key doesn't exist, generate new key pair
	if noKeys {
		// Generate new key pair
		cfg.Private, cfg.Public, err = GenerateRSAKeyPair(cfg.KeyBits)
		if err != nil {
			return nil, fmt.Errorf("getConfig(): generating RSA key pair: %w", err)
		}
		err = os.WriteFile("private.pem", cfg.Private, 0644)
		if err != nil {
			return nil, fmt.Errorf("getConfig(): writing private key to file: %w", err)
		}
		err = os.WriteFile("public.pem", cfg.Public, 0644)
		if err != nil {
			return nil, fmt.Errorf("getConfig(): writing public key to file: %w", err)
		}
	}

	// Set agentID to public key fingerprint
	cfg.AgentID, err = generateFingerprint(cfg.Public)
	if err != nil {
		return nil, fmt.Errorf("loadKeyPair(): generating public key fingerprint: %w", err)
	}

	// TODO: Check for key pair and assign agentID to pub key fingerprint
	cfg.AgentID = "mmillsap" // TODO: Temporary hardcoded for testing

	return &cfg, nil
}

// loadKeyPair() loads RSA signing key pair if the exist, or generates them if they don't
func loadKeyPair() error {
	// Check if keys exist
	var noKeys bool
	var err error
	cfg.Private, err = os.ReadFile("private.pem")
	if err != nil {
		noKeys = true
	}
	cfg.Public, err = os.ReadFile("public.pem")
	if err != nil {
		noKeys = true
	}

	// If either key doesn't exist, generate new key pair
	if noKeys {
		// Generate new key pair
		cfg.Private, cfg.Public, err = GenerateRSAKeyPair(cfg.KeyBits)
		if err != nil {
			return fmt.Errorf("getConfig(): generating RSA key pair: %w", err)
		}
		err = os.WriteFile("private.pem", cfg.Private, 0644)
		if err != nil {
			return fmt.Errorf("getConfig(): writing private key to file: %w", err)
		}
		err = os.WriteFile("public.pem", cfg.Public, 0644)
		if err != nil {
			return fmt.Errorf("getConfig(): writing public key to file: %w", err)
		}
	}

	// Set agentID to public key fingerprint
	cfg.AgentID, err = generateFingerprint(cfg.Public)
	if err != nil {
		return fmt.Errorf("loadKeyPair(): generating public key fingerprint: %w", err)
	}

	return nil
}

// GenerateRSAKeyPair generates an RSA key pair of the given bit size
func GenerateRSAKeyPair(bits int) (privateKeyPEM []byte, publicKeyPEM []byte, err error) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("rsa.GenerateKey: %w", err)
	}

	// ----- Encode private key to PEM (PKCS#1) -----
	privDER := x509.MarshalPKCS1PrivateKey(priv)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	}
	privateKeyPEM = pem.EncodeToMemory(privBlock)

	// ----- Encode public key to PEM (PKIX) -----
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("x509.MarshalPKIXPublicKey: %w", err)
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	}
	publicKeyPEM = pem.EncodeToMemory(pubBlock)

	return privateKeyPEM, publicKeyPEM, nil
}

// generateFingerprint generates a SHA-256 fingerprint of a PEM-encoded public key
func generateFingerprint(pemKey []byte) (string, error) {
	// Decode PEM block
	block, _ := pem.Decode(pemKey)
	if block == nil || block.Type != "PUBLIC KEY" {
		return "", fmt.Errorf("invalid PEM public key")
	}

	// Parse public key (to ensure it's valid)
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}

	// Re-encode to DER for hashing
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(derBytes)

	// Encode as base64 (like OpenSSH SHA256 fingerprint)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}
