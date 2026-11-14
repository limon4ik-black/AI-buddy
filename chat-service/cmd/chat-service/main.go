package main

import (
	"github.com/limon4ik-black/AI-buddy/chat-service/chat"
	"github.com/limon4ik-black/AI-buddy/shared/platform/server"
)

func main() { server.Run("chat", chat.NewApplication) }
