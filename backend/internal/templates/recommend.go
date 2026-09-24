// Package templates implements the template library: CRUD lives in the HTTP
// handlers (repo access is simple enough not to need its own service layer),
// and this file holds the recommendation ranking used by Flow B (spec §8).
package templates

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"armss-poster-studio/internal/agents"
)

var (
	wordPattern = regexp.MustCompile(`[a-z0-9]+`)
	stopwords   = map[string]bool{
		"a": true, "an": true, "the": true, "for": true, "of": true, "and": true,
		"to": true, "in": true, "on": true, "with": true, "is": true, "at": true,
		"our": true, "we": true, "this": true, "that": true, "poster": true,
	}
)

// Candidate is the subset of a template needed for ranking — a cheap,
// deterministic scorer, not a vector search (spec §8: "the library will hold
// tens of items, not thousands").
type Candidate struct {
	TemplateID   string
	Title        string
	OccasionTags []string
	Season       string
}

// Recommendation is one ranked result with a human-readable reason built from
// exactly which signals matched, so it's explainable without an LLM call.
type Recommendation struct {
	TemplateID string
	Score      int
	Reason     string
}

// Rank scores every candidate against free text query using weighted overlap
// (occasion tag ×3, season ×2, title word ×1) and returns the top n with a
// score greater than zero, highest first. This is spec §8's "step 1": cheap,
// deterministic, and always run before any LLM fallback.
func Rank(query string, candidates []Candidate, n int) []Recommendation {
	queryWords := tokenize(query)
	if len(queryWords) == 0 {
		return nil
	}

	var results []Recommendation
	for _, c := range candidates {
		score := 0
		var reasons []string

		for _, tag := range c.OccasionTags {
			if matchesAny(tokenize(tag), queryWords) {
				score += 3
				reasons = append(reasons, "matches occasion \""+tag+"\"")
			}
		}
		if c.Season != "" && matchesAny(tokenize(c.Season), queryWords) {
			score += 2
			reasons = append(reasons, "matches season \""+c.Season+"\"")
		}
		titleMatches := intersect(tokenize(c.Title), queryWords)
		if len(titleMatches) > 0 {
			score += len(titleMatches)
			reasons = append(reasons, "title shares \""+strings.Join(titleMatches, ", ")+"\"")
		}

		if score > 0 {
			reason := "Close match on your description"
			if len(reasons) > 0 {
				reason = strings.Join(dedupe(reasons), "; ")
			}
			results = append(results, Recommendation{TemplateID: c.TemplateID, Score: score, Reason: reason})
		}
	}

	sort.SliceStable(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > n {
		results = results[:n]
	}
	return results
}

// Recommend is the full spec §8 pipeline: run the cheap ranker first, and
// only fall back to one LLM call if it returned fewer than `want` results
// (and only if the Anthropic client is enabled). A failed or unavailable LLM
// fallback degrades to the step-1 results rather than failing the request —
// recommendations are a non-blocking aid, not a required step (spec §3 Flow B
// lets the user pick "no reference" regardless).
func Recommend(ctx context.Context, client *agents.Client, query string, candidates []Candidate, want int) []Recommendation {
	results := Rank(query, candidates, want)
	if len(results) >= want || !client.Enabled() {
		return results
	}

	llmResults, err := RankWithLLM(ctx, client, query, candidates, want)
	if err != nil {
		return results
	}

	seen := make(map[string]bool, len(results))
	for _, r := range results {
		seen[r.TemplateID] = true
	}
	for _, r := range llmResults {
		if len(results) >= want {
			break
		}
		if !seen[r.TemplateID] {
			results = append(results, r)
			seen[r.TemplateID] = true
		}
	}
	return results
}

func tokenize(s string) []string {
	words := wordPattern.FindAllString(strings.ToLower(s), -1)
	out := make([]string, 0, len(words))
	for _, w := range words {
		if !stopwords[w] {
			out = append(out, w)
		}
	}
	return out
}

func matchesAny(a, b []string) bool {
	return len(intersect(a, b)) > 0
}

func intersect(a, b []string) []string {
	set := make(map[string]bool, len(b))
	for _, w := range b {
		set[w] = true
	}
	var out []string
	seen := map[string]bool{}
	for _, w := range a {
		if set[w] && !seen[w] {
			out = append(out, w)
			seen[w] = true
		}
	}
	return out
}

func dedupe(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, it := range items {
		if !seen[it] {
			seen[it] = true
			out = append(out, it)
		}
	}
	return out
}
