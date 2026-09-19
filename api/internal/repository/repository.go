// Package repository реализует паттерн Repository: интерфейс доступа к задачам
// и две реализации — PostgreSQL (рабочая) и in-memory (для тестов).
package repository

import (
	"context"
	"errors"

	"task-manager/api/internal/domain"
)

// ErrNotFound возвращается, если задачи с указанным id нет.
var ErrNotFound = errors.New("task not found")

// TaskRepository абстрагирует хранилище задач от бизнес-логики.
type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetAll(ctx context.Context) ([]domain.Task, error)
	GetByID(ctx context.Context, id int) (domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id int) error
}
