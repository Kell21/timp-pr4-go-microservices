package repository

import (
	"context"
	"database/sql"
	"errors"

	"task-manager/api/internal/domain"
)

// PostgresTaskRepository хранит задачи в таблице tasks PostgreSQL.
type PostgresTaskRepository struct {
	db *sql.DB
}

// NewPostgresTaskRepository создаёт репозиторий поверх открытого пула соединений.
func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	const q = `INSERT INTO tasks (title, done) VALUES ($1, $2) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, q, task.Title, task.Done).Scan(&task.ID, &task.CreatedAt)
}

func (r *PostgresTaskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks ORDER BY id`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *PostgresTaskRepository) GetByID(ctx context.Context, id int) (domain.Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks WHERE id = $1`
	var t domain.Task
	err := r.db.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, ErrNotFound
	}
	return t, err
}

func (r *PostgresTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	const q = `UPDATE tasks SET title = $1, done = $2 WHERE id = $3`
	res, err := r.db.ExecContext(ctx, q, task.Title, task.Done, task.ID)
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func (r *PostgresTaskRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return mustAffect(res)
}

func mustAffect(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
