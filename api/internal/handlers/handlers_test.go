package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-manager/api/internal/domain"
	"task-manager/api/internal/repository"
	"task-manager/api/internal/service"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	svc := service.NewTaskService(repository.NewMemoryTaskRepository())
	srv := httptest.NewServer(New(svc).Routes())
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestAPI_CreateListUpdateDelete(t *testing.T) {
	srv := newServer(t)

	resp := do(t, http.MethodPost, srv.URL+"/api/tasks", `{"title":"Сдать ПР4"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d", resp.StatusCode)
	}
	var created domain.Task
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID != 1 || created.Title != "Сдать ПР4" || created.Done {
		t.Fatalf("created = %+v", created)
	}

	resp = do(t, http.MethodPut, srv.URL+"/api/tasks/1", `{"done":true}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d", resp.StatusCode)
	}

	resp = do(t, http.MethodGet, srv.URL+"/api/tasks", "")
	var list []domain.Task
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || !list[0].Done {
		t.Fatalf("list = %+v", list)
	}

	if resp = do(t, http.MethodDelete, srv.URL+"/api/tasks/1", ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", resp.StatusCode)
	}
	if resp = do(t, http.MethodGet, srv.URL+"/api/tasks/1", ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete status = %d", resp.StatusCode)
	}
}

func TestAPI_ErrorStatuses(t *testing.T) {
	srv := newServer(t)
	cases := []struct {
		name, method, path, body string
		want                     int
	}{
		{"empty title", http.MethodPost, "/api/tasks", `{"title":"  "}`, http.StatusUnprocessableEntity},
		{"broken json", http.MethodPost, "/api/tasks", `{"title":`, http.StatusBadRequest},
		{"unknown field", http.MethodPost, "/api/tasks", `{"name":"x"}`, http.StatusBadRequest},
		{"bad id", http.MethodGet, "/api/tasks/abc", "", http.StatusBadRequest},
		{"missing task", http.MethodPut, "/api/tasks/99", `{"done":true}`, http.StatusNotFound},
		{"wrong method", http.MethodPatch, "/api/tasks/1", "", http.StatusMethodNotAllowed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if resp := do(t, c.method, srv.URL+c.path, c.body); resp.StatusCode != c.want {
				t.Fatalf("status = %d; want %d", resp.StatusCode, c.want)
			}
		})
	}
}

func TestAPI_Health(t *testing.T) {
	srv := newServer(t)
	if resp := do(t, http.MethodGet, srv.URL+"/health", ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", resp.StatusCode)
	}
}
