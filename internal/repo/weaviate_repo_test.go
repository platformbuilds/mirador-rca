package repo

import (
	"context"
	"testing"
	"time"

	"github.com/miradorstack/mirador-rca/internal/models"
)

func TestStoreCorrelationNoEndpoint(t *testing.T) {
	r := NewWeaviateRepo("", "", time.Second)
	corr := models.CorrelationResult{CorrelationID: "corr-1", IncidentID: "incident-1", CreatedAt: time.Now()}
	if err := r.StoreCorrelation(context.Background(), "tenant", corr); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestStorePatternsNoEndpoint(t *testing.T) {
	r := NewWeaviateRepo("", "", time.Second)
	patterns := []models.FailurePattern{{ID: "p1", Name: "pattern", LastSeen: time.Now()}}
	if err := r.StorePatterns(context.Background(), "tenant", patterns); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestStoreFeedbackNoEndpoint(t *testing.T) {
	r := NewWeaviateRepo("", "", time.Second)
	fb := models.Feedback{TenantID: "tenant", CorrelationID: "corr", Correct: true, SubmittedAt: time.Now()}
	if err := r.StoreFeedback(context.Background(), fb); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestListCorrelationsSynthetic(t *testing.T) {
	r := NewWeaviateRepo("", "", time.Second)
	resp, err := r.ListCorrelations(context.Background(), models.ListCorrelationsRequest{TenantID: "tenant", Service: "checkout"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Correlations) == 0 {
		t.Fatalf("expected synthetic correlations")
	}
}
