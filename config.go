package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
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
	KeyPath       string `mapstructure:"key_path"`
	Private       []byte
	Public        []byte
}

const (
	pubKeyFile  = "public.pem"
	privKeyFile = "private.pem"
	envPrefix   = "PROV"
)

// LoadConfig() loads configuration from JSON file or environment variables,
// with environment variables taking precedence.
func LoadConfig(configFile string) (*Config, error) {

	v := viper.New()

	// Set defaults
	v.SetDefault("key_path", "keys")
	v.SetDefault("key_bits", 2048)
	v.SetDefault("jwt_exp_minutes", 5)
	v.SetDefault("heartbeat", 5)

	// Read config from file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			//return nil, fmt.Errorf("LoadConfig(): error reading config file: %w", err)
			log.Printf("LoadConfig(): error reading config file: %v", err)
		}
	}

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("LoadConfig(): unable to decode config into struct: %w", err)
	}

	// Check if keys exist & generate new key pair if either key doesn't exist
	var noKeys bool
	var err error
	cfg.Private, err = os.ReadFile(cfg.KeyPath + `/` + privKeyFile)
	if err != nil {
		log.Println("Private key not found")
		noKeys = true
	}
	cfg.Public, err = os.ReadFile(cfg.KeyPath + `/` + pubKeyFile)
	if err != nil {
		log.Println("Public key not found")
		noKeys = true
	}
	if noKeys {
		// Generate new key pair
		log.Println("Generating new RSA key pair")
		cfg.Private, cfg.Public, err = GenerateRSAKeyPair(cfg.KeyBits)
		if err != nil {
			return nil, fmt.Errorf("LoadConfig(): generating RSA key pair: %w", err)
		}
		// create keys directory if it doesn't exist
		err = os.MkdirAll(cfg.KeyPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("LoadConfig(): creating keys directory: %w", err)
		}
		err = os.WriteFile(cfg.KeyPath+`/`+privKeyFile, cfg.Private, 0644)
		if err != nil {
			return nil, fmt.Errorf("LoadConfig(): writing private key to file: %w", err)
		}
		err = os.WriteFile(cfg.KeyPath+`/`+pubKeyFile, cfg.Public, 0644)
		if err != nil {
			return nil, fmt.Errorf("LoadConfig(): writing public key to file: %w", err)
		}
	}

	// Set agentID to public key fingerprint
	cfg.AgentID, err = generateFingerprint(cfg.Public)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig(): generating public key fingerprint: %w", err)
	}

	return &cfg, nil
}

// GenerateRSAKeyPair() generates RSA key pair of the given bit size
func GenerateRSAKeyPair(bits int) (privateKeyPEM []byte, publicKeyPEM []byte, err error) {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("rsa.GenerateKey: %w", err)
	}

	// encode private key to PEM (PKCS#1)
	privDER := x509.MarshalPKCS1PrivateKey(priv)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	}
	privateKeyPEM = pem.EncodeToMemory(privBlock)

	// encode public key to PEM (PKIX)
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

// generateFingerprint() generates a SHA-256 fingerprint from PEM-encoded public key
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
