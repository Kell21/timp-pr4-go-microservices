// Package client — HTTP-клиент Web Service к REST API задач.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Task — задача в том виде, в каком её возвращает API Service.
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrNotFound — API ответил 404.
var ErrNotFound = errors.New("task not found")

// APIError — любой другой неуспешный ответ API.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d: %s", e.Status, e.Message)
}

// Client обращается к API Service по базовому адресу (например, http://api:8080).
type Client struct {
	baseURL string
	http    *http.Client
}

// New создаёт клиент с таймаутом запросов.
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

// List возвращает все задачи.
func (c *Client) List(ctx context.Context) ([]Task, error) {
	var tasks []Task
	err := c.do(ctx, http.MethodGet, "/api/tasks", nil, &tasks)
	return tasks, err
}

// Create добавляет задачу.
func (c *Client) Create(ctx context.Context, title string) (Task, error) {
	var task Task
	err := c.do(ctx, http.MethodPost, "/api/tasks", map[string]string{"title": title}, &task)
	return task, err
}

// SetDone меняет статус задачи (PUT /api/tasks/{id}).
func (c *Client) SetDone(ctx context.Context, id int, done bool) (Task, error) {
	var task Task
	err := c.do(ctx, http.MethodPut, fmt.Sprintf("/api/tasks/%d", id), map[string]bool{"done": done}, &task)
	return task, err
}

// Delete удаляет задачу.
func (c *Client) Delete(ctx context.Context, id int) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/tasks/%d", id), nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		buf, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return &APIError{Status: resp.StatusCode, Message: e.Error}
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
