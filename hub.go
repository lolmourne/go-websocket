// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lolmourne/go-websocket/model"
	"github.com/lolmourne/go-websocket/resource/chat"
	"github.com/lolmourne/go-websocket/resource/user"
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

	chatRsc chat.IResource
	userRsc user.IResource
}

func NewHub(roomID int64, subscriber pubsub.RedisPubsub, redisClient *redis.Client, chatRsc chat.IResource, userRsc user.IResource) *Hub {
	h := &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		roomID:      roomID,
		redisClient: redisClient,
		chatRsc:     chatRsc,
		userRsc:     userRsc,
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

			year, month, day := time.Now().Date()
			startTime := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			endDate := startTime.Add(time.Duration(24) * time.Hour)

			chats := h.chatRsc.GetChatsByRoomByDate(h.roomID, startTime, endDate)
			for _, chat := range chats {
				userChat := h.userRsc.GetUserByID(chat.UserID)
				if userChat == nil {
					log.Println("User Nil")
					continue
				}

				if userChat.ProfilePic == "" {
					userChat.ProfilePic = "https://i.imgur.com/cINvch3.png"
				}

				msgChat := model.Message{
					UserID:     chat.UserID,
					ProfilePic: userChat.ProfilePic,
					UserName:   userChat.Username,
					Msg:        chat.Message,
				}

				msgJson, err := json.Marshal(msgChat)
				if err != nil {
					log.Println(err)
					continue
				}
				client.send <- msgJson
			}

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
