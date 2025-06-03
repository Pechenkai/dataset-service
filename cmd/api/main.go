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

	switch os.Getenv("MODE") {
	case "cli":
		if err := app.RootCommand.ExecuteContext(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "gui":
		addr := fmt.Sprintf("%s:%d", app.Config.TechUI.Host, app.Config.TechUI.Port)
		log.Printf("TechUI running on %s", addr)
		if err := http.ListenAndServe(addr, app.WebHandler); err != nil {
			log.Fatalf("gui server error: %v", err)
		}
	default:
		addr := fmt.Sprintf("%s:%d", app.Config.HTTP.Host, app.Config.HTTP.Port)
		log.Printf("Swagger API running on %s", addr)
		if err := http.ListenAndServe(addr, app.HTTPHandler); err != nil {
			log.Fatalf("swagger server error: %v", err)
		}
	}
}
