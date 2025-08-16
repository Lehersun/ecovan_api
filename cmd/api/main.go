package main

import (
	"context"
	"log"

	"eco-van-api/internal/app"
)

func main() {
	ctx := context.Background()

	if err := app.Run(ctx); err != nil {
		log.Fatal("Failed to run application:", err)
	}
}
