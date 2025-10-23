package main

import (
	"log"

	"places/internal/app"
)

func main() {
	application := app.NewApp()

	if err := application.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
