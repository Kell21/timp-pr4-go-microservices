// API Service: REST API задач поверх PostgreSQL.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"task-manager/api/internal/handlers"
	"task-manager/api/internal/repository"
	"task-manager/api/internal/service"
	"task-manager/api/migrations"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// connectWithRetry ждёт готовности PostgreSQL (дополняет healthcheck в docker-compose).
func connectWithRetry(dsn string, attempts int, delay time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	for i := 1; i <= attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			return db, nil
		}
		log.Printf("database is not ready (attempt %d/%d): %v", i, attempts, err)
		time.Sleep(delay)
	}
	_ = db.Close()
	return nil, err
}

func main() {
	dsn := getenv("DB_DSN", "postgres://taskuser:taskpass@localhost:5432/taskdb?sslmode=disable")
	addr := ":" + getenv("PORT", "8080")

	db, err := connectWithRetry(dsn, 15, 2*time.Second)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	if err := migrations.Apply(context.Background(), db); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Println("migrations applied")

	svc := service.NewTaskService(repository.NewPostgresTaskRepository(db))
	srv := &http.Server{
		Addr:              addr,
		Handler:           handlers.New(svc).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("API service listening on %s", addr)
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
	log.Println("API service stopped")
}
