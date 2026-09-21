package main

import (
	"log"

	"go-api-starter/internal/app"
)

func main() {
	if err := app.RunWorker(); err != nil {
		log.Fatal(err)
	}
}
