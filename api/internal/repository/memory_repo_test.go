package repository

import (
	"context"
	"errors"
	"testing"

	"task-manager/api/internal/domain"
)

func TestMemoryRepo_CreateAssignsIncrementalIDs(t *testing.T) {
	repo := NewMemoryTaskRepository()
	ctx := context.Background()

	first := &domain.Task{Title: "first"}
	second := &domain.Task{Title: "second"}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, second); err != nil {
		t.Fatal(err)
	}
	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("ids = %d, %d; want 1, 2", first.ID, second.ID)
	}
	if first.CreatedAt.IsZero() {
		t.Fatal("CreatedAt must be set")
	}

	all, _ := repo.GetAll(ctx)
	if len(all) != 2 || all[0].Title != "first" || all[1].Title != "second" {
		t.Fatalf("GetAll = %+v", all)
	}
}

func TestMemoryRepo_UpdateAndGet(t *testing.T) {
	repo := NewMemoryTaskRepository()
	ctx := context.Background()
	task := &domain.Task{Title: "write report"}
	_ = repo.Create(ctx, task)

	task.Done = true
	if err := repo.Update(ctx, task); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Done {
		t.Fatal("task must be done after update")
	}
}

func TestMemoryRepo_NotFound(t *testing.T) {
	repo := NewMemoryTaskRepository()
	ctx := context.Background()

	if _, err := repo.GetByID(ctx, 42); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID err = %v; want ErrNotFound", err)
	}
	if err := repo.Update(ctx, &domain.Task{ID: 42}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update err = %v; want ErrNotFound", err)
	}
	if err := repo.Delete(ctx, 42); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete err = %v; want ErrNotFound", err)
	}
}

func TestMemoryRepo_Delete(t *testing.T) {
	repo := NewMemoryTaskRepository()
	ctx := context.Background()
	task := &domain.Task{Title: "temp"}
	_ = repo.Create(ctx, task)

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatal(err)
	}
	if all, _ := repo.GetAll(ctx); len(all) != 0 {
		t.Fatalf("GetAll after delete = %+v", all)
	}
}
