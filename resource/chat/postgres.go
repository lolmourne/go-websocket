package chat

import (
	"log"
	"time"

	"github.com/lolmourne/go-websocket/model"
)

func (dbr *DBResource) AddChat(roomID, userID int64, message string) error {
	query := `
		INSERT INTO
			chat
		(
			room_id,
			user_id,
			message,
			created_at
		)
		VALUES
		(
			$1,
			$2,
			$3,
			$4
		)
	`

	_, err := dbr.db.Exec(query, roomID, userID, message, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (dbr *DBResource) GetChatsByRoomByDate(roomID int64, startDate time.Time, endDate time.Time) []model.Chat {
	query := `
		SELECT
			chat_id,
			room_id,
			user_id,
			message,
			created_at
		FROM 
			chat
		WHERE
			room_id = $1
		AND
			created_at >= $2
		AND
			created_at <= $3
	`

	log.Println("query var", roomID, startDate, endDate)

	chats, err := dbr.db.Queryx(query, roomID, startDate, endDate)
	if err != nil {
		return make([]model.Chat, 0)
	}

	var resultChats []model.Chat
	for chats.Next() {
		var r ChatDB
		err = chats.StructScan(&r)

		if err == nil {
			resultChats = append(resultChats, model.Chat{
				ChatID:    r.ChatID.Int64,
				Message:   r.Message.String,
				CreatedAt: r.CreatedAt,
				RoomID:    r.RoomID.Int64,
				UserID:    r.UserID.Int64,
			})
		}
	}

	return resultChats
}
