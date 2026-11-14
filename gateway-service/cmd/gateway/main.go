package main

import (
	"github.com/limon4ik-black/AI-buddy/gateway-service/gateway"
	"github.com/limon4ik-black/AI-buddy/shared/platform/server"
)

func main() { server.Run("gateway", gateway.NewApplication) }
