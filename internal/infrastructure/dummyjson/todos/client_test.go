package todos

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mcp-example/internal/todo"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	return NewClient(baseURL)
}

func TestList_MethodPathQuery(t *testing.T) {
	var capturedMethod, capturedPath string
	var capturedLimit, capturedSkip string

	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		capturedLimit = r.URL.Query().Get("limit")
		capturedSkip = r.URL.Query().Get("skip")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(todoPageDTO{
			Todos: []todoDTO{{ID: 1, Todo: "test", Completed: false, UserID: 5}},
			Total: 1, Skip: 0, Limit: 20,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	page, err := client.List(context.Background(), todo.ListParams{Limit: 20, Skip: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedMethod != "GET" {
		t.Errorf("expected GET, got %s", capturedMethod)
	}
	if capturedPath != "/todos" {
		t.Errorf("expected /todos, got %s", capturedPath)
	}
	if capturedLimit != "20" {
		t.Errorf("expected limit=20, got %s", capturedLimit)
	}
	if capturedSkip != "0" {
		t.Errorf("expected skip=0, got %s", capturedSkip)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(page.Items))
	}
}

func TestList_DTOToDomainMapping(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(todoPageDTO{
			Todos: []todoDTO{
				{ID: 1, Todo: "task one", Completed: false, UserID: 10},
				{ID: 2, Todo: "task two", Completed: true, UserID: 20},
			},
			Total: 2, Skip: 0, Limit: 2,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	page, err := client.List(context.Background(), todo.ListParams{Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if page.Total != 2 {
		t.Errorf("expected total 2, got %d", page.Total)
	}
	if page.Limit != 2 {
		t.Errorf("expected limit 2, got %d", page.Limit)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(page.Items))
	}
	if page.Items[0].Text != "task one" {
		t.Errorf("expected text 'task one', got %s", page.Items[0].Text)
	}
	if !page.Items[1].Completed {
		t.Error("expected item 1 completed")
	}
	if page.Items[1].UserID != 20 {
		t.Errorf("expected userID 20, got %d", page.Items[1].UserID)
	}
}

func TestGet_404(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.Get(context.Background(), 9999)
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestList_500(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.List(context.Background(), todo.ListParams{})
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestGet_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": "not valid json"`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.Get(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestList_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.List(ctx, todo.ListParams{})
	if err == nil {
		t.Fatal("expected error for context cancellation")
	}
}

func TestGetByUser_MethodPath(t *testing.T) {
	var capturedPath string

	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(todoPageDTO{
			Todos: []todoDTO{{ID: 1, Todo: "user task", Completed: true, UserID: 5}},
			Total: 1, Skip: 0, Limit: 1,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetByUser(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/todos/user/5" {
		t.Errorf("expected /todos/user/5, got %s", capturedPath)
	}
}
