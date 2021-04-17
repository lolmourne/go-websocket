package main

import (
	"log"

	"github.com/lolmourne/go-groupchat/client"
)

type RoomManager struct {
	hubs map[int64]*Hub
	cli  client.GroupchatClientItf
}

type RoomManagerItf interface {
	JoinRoom(roomID int64) *Hub
}

func NewRoomManager(cli client.GroupchatClientItf) RoomManagerItf {
	return &RoomManager{
		hubs: make(map[int64]*Hub),
		cli:  cli,
	}
}

func (r *RoomManager) JoinRoom(roomID int64) *Hub {
	room := r.cli.GetGroupchatRoom(roomID)
	if room != nil {
		log.Println(room.Description)
	}

	hub, ok := r.hubs[roomID]
	if !ok {
		hub = NewHub()
		r.hubs[roomID] = hub
		go r.hubs[roomID].run()
		return hub
	}
	return hub
}
