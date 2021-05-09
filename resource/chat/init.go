package chat

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lolmourne/go-websocket/model"
)

type IResource interface {
	AddChat(roomID, userID int64, message string) error
	GetChatsByRoomByDate(roomID int64, startDate time.Time, endDate time.Time) []model.Chat
}

type DBResource struct {
	db *sqlx.DB
}

func NewDBResource(db *sqlx.DB) IResource {
	return &DBResource{
		db: db,
	}
}

type ChatDB struct {
	ChatID    sql.NullInt64  `db:"chat_id"`
	UserID    sql.NullInt64  `db:"user_id"`
	Message   sql.NullString `db:"message"`
	RoomID    sql.NullInt64  `db:"room_id"`
	CreatedAt time.Time      `db:"created_at"`
}
