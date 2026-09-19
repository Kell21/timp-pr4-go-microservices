package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"task-manager/api/internal/domain"
	"task-manager/api/migrations"
)

// Интеграционный тест с реальной PostgreSQL. Выполняется, только если задана
// переменная TEST_DB_DSN (в CI её задаёт сервис-контейнер postgres).
func TestPostgresRepo_CRUD(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE tasks RESTART IDENTITY`); err != nil {
		t.Fatal(err)
	}

	repo := NewPostgresTaskRepository(db)
	task := &domain.Task{Title: "integration"}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatal(err)
	}
	if task.ID == 0 || task.CreatedAt.IsZero() {
		t.Fatalf("Create must fill id and created_at: %+v", task)
	}

	task.Done = true
	if err := repo.Update(ctx, task); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByID(ctx, task.ID)
	if err != nil || !got.Done || got.Title != "integration" {
		t.Fatalf("GetByID = %+v, %v", got, err)
	}

	all, err := repo.GetAll(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("GetAll = %+v, %v", all, err)
	}

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, task.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete err = %v; want ErrNotFound", err)
	}
}
