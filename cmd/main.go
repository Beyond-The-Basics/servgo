package main

import (
	"github.com/Mohamed-Moumni/servgo/internal/server"
)

func main() {
	instance := server.New(80)

	instance.CreateServer()
}
