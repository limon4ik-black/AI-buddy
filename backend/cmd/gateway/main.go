package main

import (
	"github.com/limon4ik-black/AI-buddy/backend/internal/gateway"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/server"
)

func main() { server.Run("gateway", gateway.NewApplication) }
