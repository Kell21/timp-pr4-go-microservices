package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"task-manager/web/internal/client"
)

type fakeAPI struct {
	tasks   []client.Task
	doneArg map[int]bool
}

func (f *fakeAPI) List(context.Context) ([]client.Task, error) { return f.tasks, nil }
func (f *fakeAPI) Create(_ context.Context, title string) (client.Task, error) {
	t := client.Task{ID: len(f.tasks) + 1, Title: title}
	f.tasks = append(f.tasks, t)
	return t, nil
}
func (f *fakeAPI) SetDone(_ context.Context, id int, done bool) (client.Task, error) {
	f.doneArg[id] = done
	return client.Task{ID: id, Done: done}, nil
}
func (f *fakeAPI) Delete(context.Context, int) error { return client.ErrNotFound }

func newTestServer(t *testing.T, api *fakeAPI) http.Handler {
	t.Helper()
	tmpl := fstest.MapFS{"index.html": {Data: []byte(
		`{{range .Tasks}}<li>{{.Title}}{{if .Done}} [done]{{end}}</li>{{end}}total={{.Total}} completed={{.Completed}}`)}}
	s, err := New(api, tmpl, fstest.MapFS{})
	if err != nil {
		t.Fatal(err)
	}
	return s.Routes()
}

func TestIndex_RendersTasks(t *testing.T) {
	api := &fakeAPI{tasks: []client.Task{{ID: 1, Title: "a", Done: true}, {ID: 2, Title: "<b>"}}}
	rec := httptest.NewRecorder()
	newTestServer(t, api).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	body := rec.Body.String()
	for _, want := range []string{"<li>a [done]</li>", "&lt;b&gt;", "total=2 completed=1"} {
		if !strings.Contains(body, want) {
			t.Errorf("body %q does not contain %q", body, want)
		}
	}
}

func TestCreate_RedirectsToIndex(t *testing.T) {
	api := &fakeAPI{}
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader("title=Купить+хлеб"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	newTestServer(t, api).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}
	if len(api.tasks) != 1 || api.tasks[0].Title != "Купить хлеб" {
		t.Fatalf("tasks = %+v", api.tasks)
	}
}

func TestSetDone_AcceptsStringID(t *testing.T) {
	// data-id из HTML приходит строкой: {"id":"5","done":true}
	api := &fakeAPI{doneArg: map[int]bool{}}
	req := httptest.NewRequest(http.MethodPost, "/tasks/done", strings.NewReader(`{"id":"5","done":true}`))
	rec := httptest.NewRecorder()
	newTestServer(t, api).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || !api.doneArg[5] {
		t.Fatalf("status = %d, doneArg = %v", rec.Code, api.doneArg)
	}
}

func TestDelete_PropagatesNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/tasks/delete", strings.NewReader(`{"id":9}`))
	rec := httptest.NewRecorder()
	newTestServer(t, &fakeAPI{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404", rec.Code)
	}
}
