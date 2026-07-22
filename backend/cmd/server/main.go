package main

import (
	"context"
	"log"

	"wealthos-backend/internal/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
