package main

import (
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/server"
	"github.com/limon4ik-black/AI-buddy/backend/internal/ragmock"
)

func main() { server.Run("ragmock", ragmock.NewApplication) }
