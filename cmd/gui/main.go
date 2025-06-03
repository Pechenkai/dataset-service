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

	// Берём настройки порта из config.TechUI (или выделенный блок в config GUI)
	addr := fmt.Sprintf("%s:%d", app.Config.TechUI.Host, app.Config.TechUI.Port)
	log.Printf("GUI running on %s", addr)
	if err := http.ListenAndServe(addr, app.WebHandler); err != nil {
		log.Fatalf("gui server error: %v", err)
	}
}
