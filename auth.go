/*
 * Anasazi Precision Engineering LLC CONFIDENTIAL
 *
 * Unpublished Copyright (c) 2025 Anasazi Precision Engineering LLC. All Rights Reserved.
 *
 * Proprietary to Anasazi Precision Engineering LLC and may be covered by patents, patents
 * in process, and trade secret or copyright law. Dissemination of this information or
 * reproduction of this material is strictly forbidden unless prior written
 * permission is obtained from Anasazi Precision Engineering LLC.
 */

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
			"dev": cfg.DeviceType,
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
