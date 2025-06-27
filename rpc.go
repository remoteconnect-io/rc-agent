package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type rpc struct {
	Version int             `json:"version"`
	JobID   string          `json:"job_id"`
	SentAt  int64           `json:"sent_at"`
	Method  string          `json:"method"`
	Payload json.RawMessage `json:"payload"`
}

type sshTunnel struct {
	Remote int `json:"remote_port"`
	Local  int `json:"local_port"`
}

// heartbeat() is a Go routine that sends a heartbeat message periodically
func heartbeat(conn *websocket.Conn, agent string) {
	for {
		// Create heartbeat message
		msg := rpc{1, agent, time.Now().Unix(), "HEARTBEAT", nil}
		hb, err := json.Marshal(msg)
		if err != nil {
			log.Println("rpsRes(): ", err)
		}
		err = conn.WriteMessage(websocket.TextMessage, hb)
		if err != nil {
			log.Println("heartbeat(): ", err)
		}
		time.Sleep(time.Duration(c.Heartbeat) * time.Second)
	}
}

// rpcRes() sends an RPC response message
func rpcRes(conn *websocket.Conn, msg rpc) {
	res, err := json.Marshal(msg)
	if err != nil {
		log.Println("rpsRes(): ", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, res)
	if err != nil {
		log.Println("rpsRes(): ", err)
	}
}

// dispatch() is a Go routine used to dispatch RPC's received from server
func dispatcher(conn *websocket.Conn, rpcBuffer <-chan rpc) {
	var r rpc
	for {
		// Retrieve RPC from incoming RPC buffer. Call handlers from here.
		r = <-rpcBuffer
		switch r.Method {
		case "activate_amt": // Fake
			log.Println("dispatcher(): activate_amt received")
			msg := rpc{1, r.JobID, time.Now().Unix(), "FAILED", nil}
			rpcRes(conn, msg)
		case "create_tunnel":
			var sshPort sshTunnel
			err := json.Unmarshal(r.Payload, &sshPort)
			if err != nil {
				log.Println("dispatcher(): create_tunnel unmarshalling:", err)
			}
			fmt.Printf("Create tunnel, remote_port: %d, local_port: %d\n", sshPort.Remote, sshPort.Local)
			go createTunnel(sshPort.Local, sshPort.Remote)
		default:
			log.Println("dispatcher(): undefined RPC method: ", r.Method)
		}
	}
}
