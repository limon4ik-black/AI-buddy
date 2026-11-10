package chat

import (
	"time"

	"github.com/limon4ik-black/AI-buddy/backend/internal/contracts"
)

type Chat struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId" db:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type Message struct {
	ID              string             `json:"id"`
	ChatID          string             `json:"chatId" db:"chat_id"`
	Role            string             `json:"role"`
	Content         string             `json:"content"`
	ClientMessageID *string            `json:"clientMessageId" db:"client_message_id"`
	ReplyTo         *string            `json:"replyTo" db:"reply_to"`
	Sources         []contracts.Source `json:"sources"`
	SuggestedAction *contracts.Action  `json:"suggestedAction" db:"suggested_action"`
	Mode            string             `json:"mode"`
	Sequence        int64              `json:"sequence"`
	CreatedAt       time.Time          `json:"createdAt" db:"created_at"`
}
