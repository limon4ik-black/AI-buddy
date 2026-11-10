package main

import (
	"github.com/limon4ik-black/AI-buddy/backend/internal/info"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/server"
)

func main() { server.Run("info", info.NewApplication) }
