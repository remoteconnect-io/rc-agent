package main

import (
	"encoding/base64"
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

	// Retrieve key pair if exist, or create if they don't
	_, err = os.Stat(privFile)

	privPem, err = os.ReadFile(privFile)
	if err != nil {
		return "", fmt.Errorf("getSignedToken(): reading private key file: %w", err)
	}
	pubPem, err = os.ReadFile(pubFile)
	if err != nil {
		return "", fmt.Errorf("getSignedToken(): reading private key file: %w", err)
	}

	// Create signed JWT
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privPem)
	if err != nil {
		return "", fmt.Errorf("createToken(): parse %w", err)
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"agentID": agentID,
			"exp":     time.Now().Add(time.Second * time.Duration(cfg.JwtExpMinutes)).Unix(),
			"iss":     base64.StdEncoding.EncodeToString(pubPem),
		}).SignedString(privKey)
	if err != nil {
		return "", fmt.Errorf("getSignedToken(): sign token %w", err)
	}

	return token, nil
}
