package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Фейковый API Service проверяет, какие запросы отправляет клиент.
func TestClient_SetDoneSendsPut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/tasks/3" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]bool
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !body["done"] {
			t.Errorf("body = %v; want done=true", body)
		}
		_ = json.NewEncoder(w).Encode(Task{ID: 3, Title: "x", Done: true})
	}))
	defer srv.Close()

	task, err := New(srv.URL).SetDone(context.Background(), 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if !task.Done || task.ID != 3 {
		t.Fatalf("task = %+v", task)
	}
}

func TestClient_ErrorMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"error":"title must not be empty"}`))
		}
	}))
	defer srv.Close()
	c := New(srv.URL)

	if err := c.Delete(context.Background(), 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete err = %v; want ErrNotFound", err)
	}
	_, err := c.Create(context.Background(), "")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 422 || apiErr.Message != "title must not be empty" {
		t.Fatalf("Create err = %v", err)
	}
}
