package dto

import (
	"encoding/json"
	"time"
)

type CreateProjectRequest struct {
	Text         string   `json:"text"`
	TemplateID   *string  `json:"templateId,omitempty"`
	UploadIDs    []string `json:"uploadIds,omitempty"`
	LogoUploadID *string  `json:"logoUploadId,omitempty"`
	FooterText   *string  `json:"footerText,omitempty"`
}

type RerunBriefRequest struct {
	ExtraDetail string `json:"extraDetail"`
}

type RerunArtDirectionRequest struct {
	ExcludedObjects []string `json:"excludedObjects"`
}

type ProjectSummary struct {
	ID         string    `json:"id"`
	Title      *string   `json:"title,omitempty"`
	InputText  string    `json:"inputText"`
	TemplateID *string   `json:"templateId,omitempty"`
	State      string    `json:"state"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// StageOutputView is one entry in ProjectDetail.Stages — the current
// non-superseded row for that stage, so the UI always renders the latest state.
type StageOutputView struct {
	ID        string          `json:"id"`
	Payload   json.RawMessage `json:"payload"`
	Model     *string         `json:"model,omitempty"`
	TokensIn  *int            `json:"tokensIn,omitempty"`
	TokensOut *int            `json:"tokensOut,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

// RenderView's Stale flag is computed at read time (render.createdAt vs. the
// latest non-superseded art stage's createdAt) rather than stored — see
// pipeline design notes. The UI dims a render when Stale is true.
type RenderView struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Provider  *string   `json:"provider,omitempty"`
	URL       string    `json:"url"`
	Prompt    *string   `json:"prompt,omitempty"`
	CostCents *int      `json:"costCents,omitempty"`
	Stale     bool      `json:"stale"`
	CreatedAt time.Time `json:"createdAt"`
}

type ProjectDetail struct {
	ID         string                     `json:"id"`
	Title      *string                    `json:"title,omitempty"`
	InputText  string                     `json:"inputText"`
	TemplateID *string                    `json:"templateId,omitempty"`
	LogoURL    *string                    `json:"logoUrl,omitempty"`
	FooterText *string                    `json:"footerText,omitempty"`
	State      string                     `json:"state"`
	CreatedAt  time.Time                  `json:"createdAt"`
	UpdatedAt  time.Time                  `json:"updatedAt"`
	Stages     map[string]StageOutputView `json:"stages"`
	Renders    []RenderView               `json:"renders"`
}
