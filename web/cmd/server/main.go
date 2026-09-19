// Web Service: HTML-интерфейс Task Manager, обращается к API Service по HTTP.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-manager/web/internal/client"
	"task-manager/web/internal/ui"
	"task-manager/web/static"
	"task-manager/web/templates"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	apiURL := getenv("API_URL", "http://localhost:8081")
	addr := ":" + getenv("PORT", "8080")

	srvUI, err := ui.New(client.New(apiURL), templates.FS, static.FS)
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           srvUI.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Web service listening on %s, API_URL=%s", addr, apiURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("Web service stopped")
}
