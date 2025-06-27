package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

var c Config

func init() {
	err := getConfig(&c, "./config/config.json")
	if err != nil {
		log.Fatalf("error getting configuration settings: %v", err)
	}
}

func main() {
	authTkn, err := getSignedToken(c.AgentID)
	if err != nil {
		log.Fatalln("main(): getting authorization token: ", err)
	}

	wsHeader := http.Header{}
	wsHeader.Set("Authorization", authTkn)
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", c.ServerIP, c.ServerPort),
		Path:   "/agent/v1/ws"}

	// Connect to server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), wsHeader)
	if err != nil {
		log.Fatal("main(): websocket connection:", err)
	}
	defer conn.Close()

	// Launch heartbeat routine
	go heartbeat(conn, c.AgentID)

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
