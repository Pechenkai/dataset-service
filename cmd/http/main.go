package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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

	addr := fmt.Sprintf("%s:%d", app.Config.HTTP.Host, app.Config.HTTP.Port)
	log.Printf("Swagger API running on %s", addr)
	if err := http.ListenAndServe(addr, app.HTTPHandler); err != nil {
		log.Fatalf("api server error: %v", err)
	}
}
