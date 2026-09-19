// Package service содержит бизнес-логику работы с задачами.
package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"task-manager/api/internal/domain"
	"task-manager/api/internal/repository"
)

// MaxTitleLen — максимальная длина названия задачи (в символах).
const MaxTitleLen = 200

var (
	// ErrEmptyTitle — название пустое или состоит из пробелов.
	ErrEmptyTitle = errors.New("title must not be empty")
	// ErrTitleTooLong — название длиннее MaxTitleLen символов.
	ErrTitleTooLong = errors.New("title is too long")
	// ErrNotFound пробрасывается из репозитория.
	ErrNotFound = repository.ErrNotFound
)

// TaskService проверяет входные данные и делегирует хранение репозиторию.
type TaskService struct {
	repo repository.TaskRepository
}

// NewTaskService создаёт сервис с указанным репозиторием.
func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func normalizeTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ErrEmptyTitle
	}
	if utf8.RuneCountInString(title) > MaxTitleLen {
		return "", ErrTitleTooLong
	}
	return title, nil
}

// Create добавляет новую невыполненную задачу.
func (s *TaskService) Create(ctx context.Context, title string) (domain.Task, error) {
	title, err := normalizeTitle(title)
	if err != nil {
		return domain.Task{}, err
	}
	task := domain.Task{Title: title}
	if err := s.repo.Create(ctx, &task); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

// List возвращает все задачи.
func (s *TaskService) List(ctx context.Context) ([]domain.Task, error) {
	return s.repo.GetAll(ctx)
}

// Get возвращает задачу по id.
func (s *TaskService) Get(ctx context.Context, id int) (domain.Task, error) {
	return s.repo.GetByID(ctx, id)
}

// Update изменяет название и/или статус задачи.
func (s *TaskService) Update(ctx context.Context, id int, upd domain.TaskUpdate) (domain.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}
	if upd.Title != nil {
		title, err := normalizeTitle(*upd.Title)
		if err != nil {
			return domain.Task{}, err
		}
		task.Title = title
	}
	if upd.Done != nil {
		task.Done = *upd.Done
	}
	if err := s.repo.Update(ctx, &task); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

// Delete удаляет задачу.
func (s *TaskService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
