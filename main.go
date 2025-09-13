package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

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
	fmt.Printf("Config: %s\n", cfg.AgentID)
	if true {
		log.Fatalln(cfg)
	}
	// End TODO

	authTkn, err := getSignedToken(cfg.AgentID)
	if err != nil {
		log.Fatalln("main(): getting authorization token: ", err)
	}

	wsHeader := http.Header{}
	wsHeader.Set("Authorization", authTkn)
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", cfg.ServerIP, cfg.ServerPort),
		Path:   "/agent/v1/ws"}

	// Connect to server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), wsHeader)
	if err != nil {
		log.Fatal("main(): websocket connection:", err)
	}
	defer conn.Close()

	// Launch heartbeat routine
	go heartbeat(conn, cfg.AgentID)

	// Launch dispatcher
	rpcBuffer := make(chan rpc)
	go dispatcher(conn, rpcBuffer)

	// Continuously check for incoming messages
	var (
		msgRPC  rpc
		message []byte
	)
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

		// Send RPC to dispatcher
		rpcBuffer <- msgRPC
	}
}
