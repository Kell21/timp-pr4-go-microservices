// Package domain содержит модели предметной области «Task Manager».
package domain

import "time"

// Task — задача пользователя.
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskUpdate — частичное обновление задачи: nil-поля не изменяются.
type TaskUpdate struct {
	Title *string `json:"title,omitempty"`
	Done  *bool   `json:"done,omitempty"`
}
