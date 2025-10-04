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

type heartbeat struct {
	AgentID string
}

// heartbeat() is a Go routine that sends a heartbeat message periodically
func sendHeartbeat(conn *websocket.Conn, agent string) {
	for {
		// Create heartbeat message
		payload := heartbeat{cfg.AgentID}
		encPayload, err := json.Marshal(payload)
		if err != nil {
			log.Println("sendHeartbeat(): payload encoding: ", err)
		}
		msg := rpc{1, agent, time.Now().Unix(), "HEARTBEAT", encPayload}
		hb, err := json.Marshal(msg)
		if err != nil {
			log.Println("sendHeartbeat(): heartbeat JSON marshaling ", err)
		}

		// Send heartbeat
		err = conn.WriteMessage(websocket.TextMessage, hb)
		if err != nil {
			log.Println("sendHeartbeat(): ", err)
		}
		time.Sleep(time.Duration(cfg.Heartbeat) * time.Minute)
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
		case "rotate_keys": // Fake
			log.Println("dispatcher(): rotate_keys received")
			msg := rpc{1, r.JobID, time.Now().Unix(), "FAILED", nil}
			rpcRes(conn, msg)
		case "provision_recipe":
			log.Println("dispatcher(): provision_recipe received")
			msg := rpc{1, r.JobID, time.Now().Unix(), "provision_ack", nil}
			rpcRes(conn, msg)

			// TODO: need to validate & do something with recipe
		default:
			log.Println("dispatcher(): undefined RPC method: ", r.Method)
		}
	}
}
