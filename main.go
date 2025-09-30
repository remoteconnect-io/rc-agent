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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var cfg *Config

func init() {
	var err error
	cfg, err = LoadConfig("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
}

func main() {
	// If device type is 'device' then get OTP & ProvAPI URL from filesystem,
	//  else generate a new OTP and use default ProvAPI URL from config
	var otp string
	var err error
	if strings.ToLower(cfg.DeviceType) == "device" {
		content, err := os.ReadFile(cfg.LocalOTPPath)
		if err != nil {
			log.Fatalln("main(): reading OTP from file: ", err)
		}
		otp = string(content)
		content, err = os.ReadFile(cfg.LocalURLPath)
		if err != nil {
			log.Fatalln("main(): reading URL from file: ", err)
		}
		cfg.serverAddr = strings.TrimSpace(string(content))
	} else {
		otp, err = generateOTP()
		if err != nil {
			log.Fatalln("main(): generating OTP: ", err)
		}
	}

	authTkn, err := getSignedJWT(otp)
	if err != nil {
		log.Fatalln("main(): getting authorization token: ", err)
	}

	wsHeader := http.Header{}
	wsHeader.Set("Authorization", authTkn)
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", cfg.serverAddr, cfg.ServerPort),
		Path:   "/agent/v1/ws"}

	// TODO: remove this test code
	log.Println(authTkn)

	// Loop until connected
	var conn *websocket.Conn
	for {
		conn, _, err = websocket.DefaultDialer.Dial(u.String(), wsHeader)
		if err == nil {
			break
		}
		log.Printf("main(): websocket connection: %v\n", err)
		time.Sleep(time.Duration(cfg.ConnTimeoutSec) * time.Second)
	}
	defer conn.Close()
	log.Printf("%s connected...\n", cfg.AgentID)

	// Launch heartbeat routine
	go sendHeartbeat(conn, cfg.AgentID)

	// Launch dispatcher routine
	rpcBuffer := make(chan rpc)
	go dispatcher(conn, rpcBuffer)

	var (
		msgRPC  rpc
		message []byte
	)

	// Continuously check for incoming messages
	for {
		_, message, err = conn.ReadMessage()
		if err != nil {
			log.Println("main(): websocket read error:", err)
			break
		}
		err = json.Unmarshal(message, &msgRPC)
		if err != nil {
			log.Println("main(): unmashal websocket message error:", err)
			break
		}

		// Send RPC to dispatcher channel
		rpcBuffer <- msgRPC
	}
}
