package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	if os.Getenv("CLI") == "1" {
		if err := app.RootCommand.ExecuteContext(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	port := app.Config.HTTP.Port
	addr := fmt.Sprintf(":%d", port)
	log.Printf("HTTP server running on %s", addr)
	if err := http.ListenAndServe(addr, app.HTTPHandler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
