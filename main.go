package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	// TODO: remove this test code
	log.Println(cfg)

	otp, err := generateOTP()
	if err != nil {
		log.Fatalln("main(): generating OTP: ", err)
	}
	authTkn, err := getSignedJWT(otp)
	if err != nil {
		log.Fatalln("main(): getting authorization token: ", err)
	}

	wsHeader := http.Header{}
	wsHeader.Set("Authorization", authTkn)
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", cfg.ServerIP, cfg.ServerPort),
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

	// Launch heartbeat routine
	go heartbeat(conn, cfg.AgentID)

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
