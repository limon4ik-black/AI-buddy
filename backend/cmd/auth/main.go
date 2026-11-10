package main

import (
	"github.com/limon4ik-black/AI-buddy/backend/internal/auth"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/server"
)

func main() { server.Run("auth", auth.NewApplication) }
