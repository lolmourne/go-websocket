// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/lolmourne/r-pipeline/pubsub"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	roomID int64

	redisClient *redis.Client
}

func NewHub(roomID int64, subscriber pubsub.RedisPubsub, redisClient *redis.Client) *Hub {
	h := &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		roomID:      roomID,
		redisClient: redisClient,
	}
	subscriber.Subscribe(fmt.Sprintf("pubsub:chat:%d", roomID), h.readRoomPubsub, true)

	return h
}

func (h *Hub) readRoomPubsub(msg string, err error) {
	if err != nil {
		log.Println(err)
		return
	}

	for client := range h.clients {
		client.send <- []byte(msg)
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			h.redisClient.Publish(context.Background(), fmt.Sprintf("pubsub:chat:%d", h.roomID), message)
		}
	}
}
