package model

import "time"

type Chat struct {
	ChatID    int64     `json:"chat_id"`
	UserID    int64     `json:"user_id"`
	Message   string    `json:"message"`
	RoomID    int64     `json:"room_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	UserID     int64  `json:"user_id"`
	UserName   string `json:"username"`
	ProfilePic string `json:"profile_pic"`
	Msg        string `json:"msg"`
}

type User struct {
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
	ProfilePic string    `json:"profile_pic"`
}
