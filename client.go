// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	authClient "github.com/lolmourne/go-accounts/client/userauth"
	"github.com/lolmourne/go-websocket/model"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	usr User
}

type User struct {
	UserID     int64
	ProfilePic string
	Username   string
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		if c.usr.ProfilePic == "" {
			c.usr.ProfilePic = "https://i.imgur.com/cINvch3.png"
		}

		c.hub.chatRsc.AddChat(c.hub.roomID, c.usr.UserID, string(message))

		msgObj := model.Message{
			UserID:     c.usr.UserID,
			ProfilePic: c.usr.ProfilePic,
			UserName:   c.usr.Username,
			Msg:        string(message),
		}

		msg, err := json.Marshal(msgObj)
		if err != nil {
			log.Println(err)
		} else {
			c.hub.broadcast <- msg
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(room RoomManagerItf, authClient authClient.ClientItf, w http.ResponseWriter, r *http.Request) {
	roomIDStr := r.URL.Query().Get("room_id")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	authToken := r.URL.Query().Get("authToken")
	if err != nil || authToken == "" {
		log.Println("auth failed")
		return
	}

	userInfo := authClient.GetUserInfo(authToken)
	if userInfo == nil {
		log.Println("auth failed")
		w.Write([]byte("User not found"))
		return
	}

	hub := room.JoinRoom(roomID)
	if hub == nil {
		w.Write([]byte("Room not found"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	user := User{
		UserID:     userInfo.UserID,
		Username:   userInfo.Username,
		ProfilePic: userInfo.ProfilePic,
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), usr: user}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
