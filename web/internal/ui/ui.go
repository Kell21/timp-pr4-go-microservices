// Package ui — HTTP-обработчики пользовательского интерфейса Task Manager.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"task-manager/web/internal/client"
)

// TaskAPI — операции API Service, которые использует интерфейс.
// Реализуется *client.Client; в тестах подменяется фейком.
type TaskAPI interface {
	List(ctx context.Context) ([]client.Task, error)
	Create(ctx context.Context, title string) (client.Task, error)
	SetDone(ctx context.Context, id int, done bool) (client.Task, error)
	Delete(ctx context.Context, id int) error
}

// Server отдаёт HTML-страницу, статику и принимает действия пользователя.
type Server struct {
	api  TaskAPI
	tmpl *template.Template
	stat fs.FS
}

// New разбирает шаблоны из templates и создаёт сервер.
func New(api TaskAPI, templates, static fs.FS) (*Server, error) {
	tmpl, err := template.ParseFS(templates, "*.html")
	if err != nil {
		return nil, err
	}
	return &Server{api: api, tmpl: tmpl, stat: static}, nil
}

// Routes регистрирует маршруты веб-интерфейса.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(s.stat)))
	mux.HandleFunc("GET /{$}", s.index)
	mux.HandleFunc("POST /tasks", s.create)
	mux.HandleFunc("POST /tasks/done", s.setDone)
	mux.HandleFunc("POST /tasks/delete", s.remove)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

type pageData struct {
	Tasks     []client.Task
	Total     int
	Completed int
	Error     string
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	data := pageData{Error: r.URL.Query().Get("error")}
	tasks, err := s.api.List(r.Context())
	if err != nil {
		log.Printf("list tasks: %v", err)
		data.Error = "API Service недоступен: " + err.Error()
	}
	data.Tasks = tasks
	data.Total = len(tasks)
	for _, t := range tasks {
		if t.Done {
			data.Completed++
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("render: %v", err)
	}
}

// create обрабатывает обычную HTML-форму и возвращает пользователя на главную (PRG).
func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	if _, err := s.api.Create(r.Context(), r.FormValue("title")); err != nil {
		http.Redirect(w, r, "/?error="+url.QueryEscape(humanError(err)), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type doneRequest struct {
	ID   json.Number `json:"id"`
	Done bool        `json:"done"`
}

// setDone вызывается из JavaScript (fetch) при изменении чекбокса.
func (s *Server) setDone(w http.ResponseWriter, r *http.Request) {
	var req doneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(req.ID.String())
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	task, err := s.api.SetDone(r.Context(), id, req.Done)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

// remove вызывается из JavaScript (fetch) по кнопке «Удалить».
func (s *Server) remove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID json.Number `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(req.ID.String())
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := s.api.Delete(r.Context(), id); err != nil {
		writeAPIError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeAPIError(w http.ResponseWriter, err error) {
	var apiErr *client.APIError
	switch {
	case errors.Is(err, client.ErrNotFound):
		http.Error(w, "задача не найдена", http.StatusNotFound)
	case errors.As(err, &apiErr):
		http.Error(w, apiErr.Message, apiErr.Status)
	default:
		http.Error(w, "API Service недоступен", http.StatusBadGateway)
	}
}

func humanError(err error) string {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
		return "Название задачи должно содержать от 1 до 200 символов"
	}
	return "Не удалось добавить задачу: " + err.Error()
}
