package main

import (
	"github.com/limon4ik-black/AI-buddy/ragmock-service/ragmock"
	"github.com/limon4ik-black/AI-buddy/shared/platform/server"
)

func main() { server.Run("ragmock", ragmock.NewApplication) }
