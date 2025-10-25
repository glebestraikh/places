package main

import (
	"log"
	"places/internal/app"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(1) // горутины асинхронны, но не параллельны

	application := app.NewApp()

	if err := application.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
