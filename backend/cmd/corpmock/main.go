package main

import (
	"github.com/limon4ik-black/AI-buddy/backend/internal/corpmock"
	"github.com/limon4ik-black/AI-buddy/backend/internal/platform/server"
)

func main() { server.Run("corpmock", corpmock.NewApplication) }
