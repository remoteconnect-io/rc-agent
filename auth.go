package main

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getSignedJWT() returns a signed JWT
func getSignedJWT() (token string, err error) {
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(cfg.Private)
	if err != nil {
		return "", fmt.Errorf("createToken(): parse %w", err)
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"agentID": cfg.AgentID,
			"exp":     time.Now().Add(time.Second * time.Duration(cfg.JwtExpMinutes)).Unix(),
			"iss":     base64.StdEncoding.EncodeToString(cfg.Public),
		}).SignedString(privKey)
	if err != nil {
		return "", fmt.Errorf("getSignedJWT(): error signing jwt %w", err)
	}

	return token, nil
}
