package repo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchServiceGraphCachesResults(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != "/api/v1/rca/service-graph" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"edges": []map[string]any{
				{"source": "checkout", "target": "payments", "call_rate": 42.0, "error_rate": 0.01},
			},
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	cacheStub := newStubCache()
	client := NewMiradorCoreClient(srv.URL, "/metrics", "/logs", "/traces", "/api/v1/rca/service-graph", time.Second, cacheStub, time.Minute)
	client.httpClient = srv.Client()

	ctx := context.Background()
	start := time.Unix(1_700_000_000, 0)
	end := start.Add(5 * time.Minute)

	edges, err := client.FetchServiceGraph(ctx, "tenant-a", start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected one upstream request, got %d", hits)
	}
	if len(edges) != 1 || edges[0].Source != "checkout" {
		t.Fatalf("unexpected response: %+v", edges)
	}

	cached, err := client.FetchServiceGraph(ctx, "tenant-a", start, end)
	if err != nil {
		t.Fatalf("unexpected cached error: %v", err)
	}
	if hits != 1 {
		t.Fatalf("cache miss triggered network call; hits=%d", hits)
	}
	if len(cached) != 1 || cached[0].Target != "payments" {
		t.Fatalf("unexpected cached payload: %+v", cached)
	}
}
