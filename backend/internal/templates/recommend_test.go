package templates

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"armss-poster-studio/internal/agents"
)

func sampleCandidates() []Candidate {
	return []Candidate{
		{TemplateID: "diwali-1", Title: "Diwali Lights Celebration", OccasionTags: []string{"diwali", "festival"}, Season: "autumn"},
		{TemplateID: "pongal-1", Title: "Pongal Harvest Greeting", OccasionTags: []string{"pongal", "harvest"}, Season: "winter"},
		{TemplateID: "hiring-1", Title: "We're Hiring Announcement", OccasionTags: []string{"hiring", "jobs"}, Season: ""},
		{TemplateID: "independence-1", Title: "Independence Day Tribute", OccasionTags: []string{"independence-day"}, Season: "monsoon"},
	}
}

func TestRank_OrdersByWeightedOverlap(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantFirst string
		wantCount int
	}{
		{name: "occasion tag match wins", query: "diwali lights celebration this year", wantFirst: "diwali-1", wantCount: 1},
		{name: "season plus title match", query: "harvest season pongal wishes", wantFirst: "pongal-1", wantCount: 1},
		{name: "hiring keyword", query: "we are hiring for new roles", wantFirst: "hiring-1", wantCount: 1},
		{name: "no overlap at all", query: "completely unrelated corporate retreat agenda", wantFirst: "", wantCount: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Rank(tt.query, sampleCandidates(), 3)
			if len(results) != tt.wantCount {
				t.Fatalf("expected %d results, got %d: %+v", tt.wantCount, len(results), results)
			}
			if tt.wantCount > 0 && results[0].TemplateID != tt.wantFirst {
				t.Errorf("expected top result %q, got %q", tt.wantFirst, results[0].TemplateID)
			}
		})
	}
}

func TestRank_CapsAtN(t *testing.T) {
	results := Rank("diwali pongal hiring independence festival greeting", sampleCandidates(), 2)
	if len(results) != 2 {
		t.Fatalf("expected exactly 2 results (capped), got %d", len(results))
	}
}

func TestRank_EmptyQueryReturnsNothing(t *testing.T) {
	if results := Rank("", sampleCandidates(), 3); len(results) != 0 {
		t.Fatalf("expected no results for empty query, got %+v", results)
	}
}

func TestRecommend_SkipsLLMWhenStepOneHasEnough(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError) // would fail the test if ever called
	}))
	defer srv.Close()
	client := agents.NewClient("test-key", "claude-sonnet-4-6", srv.URL, 5*time.Second)

	results := Recommend(context.Background(), client, "diwali lights celebration", sampleCandidates(), 1)
	if len(results) != 1 {
		t.Fatalf("expected 1 result from step 1 alone, got %d", len(results))
	}
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatalf("expected the LLM fallback never to be called when step 1 already has enough, got %d calls", calls)
	}
}

func TestRecommend_FallsBackToLLMWhenStepOneIsShort(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"content": [{"type": "tool_use", "input": {"rankings": [
				{"templateId": "hiring-1", "reason": "closest available match"},
				{"templateId": "independence-1", "reason": "next closest"}
			]}}],
			"stop_reason": "tool_use",
			"usage": {"input_tokens": 100, "output_tokens": 50}
		}`))
	}))
	defer srv.Close()
	client := agents.NewClient("test-key", "claude-sonnet-4-6", srv.URL, 5*time.Second)

	// A query matching only one candidate under step 1 (score-only "diwali").
	results := Recommend(context.Background(), client, "diwali", sampleCandidates(), 3)
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 LLM fallback call, got %d", calls)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results after merging step 1 + LLM fallback, got %d: %+v", len(results), results)
	}
	if results[0].TemplateID != "diwali-1" {
		t.Errorf("expected step-1 result to stay first, got %q", results[0].TemplateID)
	}
}

func TestRecommend_DisabledClientReturnsStepOneOnly(t *testing.T) {
	client := agents.NewClient("", "claude-sonnet-4-6", "http://unused.invalid", 5*time.Second)

	results := Recommend(context.Background(), client, "diwali", sampleCandidates(), 3)
	if len(results) != 1 {
		t.Fatalf("expected only the 1 step-1 result with no key configured, got %d", len(results))
	}
}
