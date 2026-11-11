package main

import (
	"github.com/limon4ik-black/AI-buddy/backend/internal/chat"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/server"
)

func main() { server.Run("chat", chat.NewApplication) }
