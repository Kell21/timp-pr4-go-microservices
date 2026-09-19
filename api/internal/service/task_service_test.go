package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"task-manager/api/internal/domain"
	"task-manager/api/internal/repository"
)

func newService() *TaskService {
	return NewTaskService(repository.NewMemoryTaskRepository())
}

func TestCreate_TrimsTitle(t *testing.T) {
	task, err := newService().Create(context.Background(), "   buy milk  ")
	if err != nil {
		t.Fatal(err)
	}
	if task.Title != "buy milk" || task.Done || task.ID == 0 {
		t.Fatalf("unexpected task %+v", task)
	}
}

func TestCreate_Validation(t *testing.T) {
	svc := newService()
	cases := []struct {
		name  string
		title string
		want  error
	}{
		{"empty", "", ErrEmptyTitle},
		{"spaces", "   \t ", ErrEmptyTitle},
		{"too long", strings.Repeat("я", MaxTitleLen+1), ErrTitleTooLong},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := svc.Create(context.Background(), c.title); !errors.Is(err, c.want) {
				t.Fatalf("err = %v; want %v", err, c.want)
			}
		})
	}
	if _, err := svc.Create(context.Background(), strings.Repeat("я", MaxTitleLen)); err != nil {
		t.Fatalf("title of exactly %d runes must be accepted: %v", MaxTitleLen, err)
	}
}

func TestUpdate_ToggleDoneKeepsTitle(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	task, _ := svc.Create(ctx, "prepare slides")

	done := true
	updated, err := svc.Update(ctx, task.ID, domain.TaskUpdate{Done: &done})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Done || updated.Title != "prepare slides" {
		t.Fatalf("unexpected task %+v", updated)
	}
}

func TestUpdate_RejectsEmptyTitle(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	task, _ := svc.Create(ctx, "original")

	empty := " "
	if _, err := svc.Update(ctx, task.ID, domain.TaskUpdate{Title: &empty}); !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("err = %v; want ErrEmptyTitle", err)
	}
	got, _ := svc.Get(ctx, task.ID)
	if got.Title != "original" {
		t.Fatalf("title changed to %q", got.Title)
	}
}

func TestUpdateAndDelete_NotFound(t *testing.T) {
	svc := newService()
	done := true
	if _, err := svc.Update(context.Background(), 7, domain.TaskUpdate{Done: &done}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update err = %v; want ErrNotFound", err)
	}
	if err := svc.Delete(context.Background(), 7); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete err = %v; want ErrNotFound", err)
	}
}
