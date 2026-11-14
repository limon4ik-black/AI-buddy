package main

import (
	"github.com/limon4ik-black/AI-buddy/corpmock-service/corpmock"
	"github.com/limon4ik-black/AI-buddy/shared/platform/server"
)

func main() { server.Run("corpmock", corpmock.NewApplication) }
