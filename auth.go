package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getSignedToken() returns the stored private & public keys in pem format,
// if they exist, or creates a new key pair if they don't exist.
func getSignedToken(agentID string) (token string, err error) {
	const privFile = "private.pem"
	const pubFile = "public.pem"
	var privPem, pubPem []byte

	// Retrieve key pair, or create if they don't exist
	_, err = os.Stat(c.KeyPath + `/` + privFile)
	if os.IsNotExist(err) {
		// Generate private & public keys
		privKey, err := rsa.GenerateKey(rand.Reader, c.KeySize)
		if err != nil {
			return "", fmt.Errorf("getSignedToken() generating private pem to file: %w", err)
		}
		pubKey := &privKey.PublicKey

		// Convert private & public keys to pem format
		privBlock := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   x509.MarshalPKCS1PrivateKey(privKey),
		}
		privPem = pem.EncodeToMemory(&privBlock)
		err = os.MkdirAll(c.KeyPath, 0644)
		if err != nil {
			return "", fmt.Errorf("getSignedToken() creating private pem directory: %w", err)
		}
		err = os.WriteFile(c.KeyPath+`/`+privFile, privPem, 0644)
		if err != nil {
			return "", fmt.Errorf("getSignedToken() writing private pem to file: %w", err)
		}
		pubBlock := pem.Block{
			Type:    "RSA PUBLIC KEY",
			Headers: nil,
			Bytes:   x509.MarshalPKCS1PublicKey(pubKey),
		}
		pubPem = pem.EncodeToMemory(&pubBlock)
		err = os.WriteFile(c.KeyPath+`/`+pubFile, pubPem, 0644)
		if err != nil {
			return "", fmt.Errorf("getSignedToken() writing public pem to file: %w", err)
		}
	} else {
		privPem, err = os.ReadFile(c.KeyPath + `/` + privFile)
		if err != nil {
			return "", fmt.Errorf("getSignedToken(): reading private key file: %w", err)
		}
		pubPem, err = os.ReadFile(c.KeyPath + `/` + pubFile)
		if err != nil {
			return "", fmt.Errorf("getSignedToken(): reading private key file: %w", err)
		}
	}

	// Create signed JWT
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privPem)
	if err != nil {
		return "", fmt.Errorf("createToken(): parse %w", err)
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"agentID": agentID,
			"exp":     time.Now().Add(time.Second * time.Duration(c.JwtExpMinutes)).Unix(),
			"iss":     base64.StdEncoding.EncodeToString(pubPem),
		}).SignedString(privKey)
	if err != nil {
		return "", fmt.Errorf("getSignedToken(): sign token %w", err)
	}

	return token, nil
}
