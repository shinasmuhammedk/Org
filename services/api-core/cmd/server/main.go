package main

import (
	"org/api-core/internal/app"
	"org/api-core/internal/server"
)

func main() {
	r := app.New()

	srv := server.New(r, ":8080")
	srv.Run()
}