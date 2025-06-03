package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"ppo/internal/di"
)

func main() {
	ctx := context.Background()

	app, err := di.Build(ctx)
	if err != nil {
		log.Fatalf("failed to build application: %v", err)
	}
	defer func() {
		if err := app.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	// Запускаем только CLI (RootCommand)
	if err := app.RootCommand.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
