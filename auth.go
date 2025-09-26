package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getSignedJWT() returns a signed JWT
func getSignedJWT(otp string) (token string, err error) {
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(cfg.Private)
	if err != nil {
		return "", fmt.Errorf("createToken(): parse %w", err)
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"sub": cfg.AgentID,
			"per": cfg.DeviceType,
			"otp": otp,
			"exp": time.Now().Add(time.Second * time.Duration(cfg.JwtExpMinutes)).Unix(),
			"iss": base64.StdEncoding.EncodeToString(cfg.Public),
		}).SignedString(privKey)
	if err != nil {
		return "", fmt.Errorf("getSignedJWT(): error signing jwt %w", err)
	}

	return token, nil
}

// GenerateOTP() returns a random 6-character alphanumeric string.
func generateOTP() (string, error) {
	const tokenCharset = "ABCDEFGHIJKLMNPQRSTUVWXYZ123456789"

	length := 6
	token := make([]byte, length)

	for i := 0; i < length; i++ {
		// pick a random index in charset
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenCharset))))
		if err != nil {
			return "", fmt.Errorf("GenerateToken(): failed to generate random number: %w\n", err)
		}
		token[i] = tokenCharset[num.Int64()]
	}

	return string(token), nil
}
