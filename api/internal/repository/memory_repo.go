package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"task-manager/api/internal/domain"
)

// MemoryTaskRepository — потокобезопасная in-memory реализация для unit-тестов.
type MemoryTaskRepository struct {
	mu     sync.Mutex
	tasks  map[int]domain.Task
	nextID int
}

// NewMemoryTaskRepository создаёт пустое in-memory хранилище.
func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{tasks: make(map[int]domain.Task), nextID: 1}
}

func (r *MemoryTaskRepository) Create(_ context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	task.ID = r.nextID
	task.CreatedAt = time.Now().UTC()
	r.tasks[task.ID] = *task
	r.nextID++
	return nil
}

func (r *MemoryTaskRepository) GetAll(_ context.Context) ([]domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tasks := make([]domain.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks, nil
}

func (r *MemoryTaskRepository) GetByID(_ context.Context, id int) (domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return domain.Task{}, ErrNotFound
	}
	return t, nil
}

func (r *MemoryTaskRepository) Update(_ context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[task.ID]; !ok {
		return ErrNotFound
	}
	r.tasks[task.ID] = *task
	return nil
}

func (r *MemoryTaskRepository) Delete(_ context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(r.tasks, id)
	return nil
}
