package main

import (
	"github.com/limon4ik-black/AI-buddy/auth-service/auth"
	"github.com/limon4ik-black/AI-buddy/shared/platform/server"
)

func main() { server.Run("auth", auth.NewApplication) }
