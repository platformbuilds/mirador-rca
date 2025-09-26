package repo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/miradorstack/mirador-rca/internal/cache"
	"github.com/miradorstack/mirador-rca/internal/models"
)

func TestStoreCorrelationNoEndpoint(t *testing.T) {
	r := NewWeaviateRepo("", "", time.Second, cache.NoopProvider{}, 0, 0)
	corr := models.CorrelationResult{CorrelationID: "corr-1", IncidentID: "incident-1", CreatedAt: time.Now()}
	if err := r.StoreCorrelation(context.Background(), "tenant", corr); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestStorePatternsNoEndpoint(t *testing.T) {
	r := NewWeaviateRepo("", "", time.Second, cache.NoopProvider{}, 0, 0)
	patterns := []models.FailurePattern{{ID: "p1", Name: "pattern", LastSeen: time.Now()}}
	if err := r.StorePatterns(context.Background(), "tenant", patterns); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestStoreFeedbackNoEndpoint(t *testing.T) {
	r := NewWeaviateRepo("", "", time.Second, cache.NoopProvider{}, 0, 0)
	fb := models.Feedback{TenantID: "tenant", CorrelationID: "corr", Correct: true, SubmittedAt: time.Now()}
	if err := r.StoreFeedback(context.Background(), fb); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestListCorrelationsSynthetic(t *testing.T) {
	r := NewWeaviateRepo("", "", time.Second, cache.NoopProvider{}, 0, 0)
	resp, err := r.ListCorrelations(context.Background(), models.ListCorrelationsRequest{TenantID: "tenant", Service: "checkout"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Correlations) == 0 {
		t.Fatalf("expected synthetic correlations")
	}
}

func TestSimilarIncidentsCachesResults(t *testing.T) {
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != "/v1/graphql" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"Get":{"CorrelationRecord":[{"correlationId":"c-1","incidentId":"inc-1","rootCause":"checkout","confidence":0.8,"affectedServices":["checkout"],"recommendations":["scale"],"createdAt":"2024-01-02T15:04:05Z"}]}}}`)
	}))
	defer server.Close()

	cacheStub := newStubCache()
	repo := NewWeaviateRepo(server.URL, "", time.Second, cacheStub, time.Minute, 0)
	repo.httpClient = server.Client()

	ctx := context.Background()
	first, err := repo.SimilarIncidents(ctx, "tenant-a", []string{"checkout", "payments"}, 2)
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected one upstream call, got %d", hits)
	}
	if len(first) != 1 || first[0].CorrelationID != "c-1" {
		t.Fatalf("unexpected correlation payload: %+v", first)
	}

	second, err := repo.SimilarIncidents(ctx, "tenant-a", []string{"payments", "checkout"}, 2)
	if err != nil {
		t.Fatalf("unexpected error on cached call: %v", err)
	}
	if hits != 1 {
		t.Fatalf("cache miss triggered network call; hits=%d", hits)
	}
	if len(second) != 1 || second[0].CorrelationID != "c-1" {
		t.Fatalf("unexpected cached payload: %+v", second)
	}
}

func TestFetchPatternsCachesResults(t *testing.T) {
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"Get":{"FailurePattern":[{"patternId":"p1","name":"error","description":"desc","services":["payments"],"anchorTemplates":[{"service":"payments","signalType":"logs","selector":"error","typicalLeadLag":1,"thresholds":10}],"prevalence":0.5,"lastSeen":"2024-01-02T15:04:05Z","quality":{"precision":0.7,"recall":0.4}}]}}}`)
	}))
	defer server.Close()
	cacheStub := newStubCache()
	repo := NewWeaviateRepo(server.URL, "", time.Second, cacheStub, time.Minute, time.Hour)
	repo.httpClient = server.Client()

	ctx := context.Background()
	first, err := repo.FetchPatterns(ctx, "tenant-a", "payments")
	if err != nil {
		t.Fatalf("unexpected error on first fetch: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected one upstream call, got %d", hits)
	}
	if len(first) != 1 || first[0].ID != "p1" {
		t.Fatalf("unexpected pattern payload: %+v", first)
	}

	second, err := repo.FetchPatterns(ctx, "tenant-a", "payments")
	if err != nil {
		t.Fatalf("unexpected error on cached fetch: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected cached response without new hit, hits=%d", hits)
	}
	if len(second) != 1 || second[0].ID != "p1" {
		t.Fatalf("unexpected cached pattern payload: %+v", second)
	}
}
