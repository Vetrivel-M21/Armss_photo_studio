package handlers

import (
	"encoding/json"
	"net/http"

	"armss-poster-studio/internal/agents"
	"armss-poster-studio/internal/http/dto"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/storage"
	"armss-poster-studio/internal/store/mysql"
	"armss-poster-studio/internal/templates"
)

const recommendationCount = 3

type RecommendationsHandler struct {
	Templates *mysql.TemplatesRepo
	Store     *storage.Store
	Agents    *agents.Client
}

func NewRecommendationsHandler(templatesRepo *mysql.TemplatesRepo, store *storage.Store, agentsClient *agents.Client) *RecommendationsHandler {
	return &RecommendationsHandler{Templates: templatesRepo, Store: store, Agents: agentsClient}
}

// Recommend implements POST /recommendations (Flow B, spec §3/§8): given free
// text with no template selected, return the top 3 templates with a reason
// each. Runs entirely without an Anthropic key when tag/keyword overlap alone
// finds 3 matches; only reaches for the LLM fallback otherwise.
func (h *RecommendationsHandler) Recommend(w http.ResponseWriter, r *http.Request) {
	var req dto.RecommendationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		respond.Error(w, http.StatusBadRequest, "Provide the occasion description in \"text\".")
		return
	}

	rows, err := h.Templates.ListActive(r.Context(), "")
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "Could not load templates to recommend from.")
		return
	}

	byID := make(map[string]mysql.Template, len(rows))
	candidates := make([]templates.Candidate, 0, len(rows))
	for _, t := range rows {
		byID[t.ID] = t
		var tags []string
		if len(t.OccasionTags) > 0 {
			_ = json.Unmarshal(t.OccasionTags, &tags)
		}
		season := ""
		if t.Season.Valid {
			season = t.Season.String
		}
		candidates = append(candidates, templates.Candidate{
			TemplateID: t.ID, Title: t.Title, OccasionTags: tags, Season: season,
		})
	}

	recs := templates.Recommend(r.Context(), h.Agents, req.Text, candidates, recommendationCount)

	out := make([]dto.RecommendationView, 0, len(recs))
	for _, rec := range recs {
		t, ok := byID[rec.TemplateID]
		if !ok {
			continue
		}
		view, err := templateToResponse(r.Context(), h.Store, t)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "Could not load recommended templates right now.")
			return
		}
		out = append(out, dto.RecommendationView{Template: view, Reason: rec.Reason})
	}
	respond.JSON(w, http.StatusOK, out)
}
