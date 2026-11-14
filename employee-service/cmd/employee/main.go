package main

import (
	"github.com/limon4ik-black/AI-buddy/employee-service/employee"
	"github.com/limon4ik-black/AI-buddy/shared/platform/server"
)

func main() { server.Run("employee", employee.NewApplication) }
