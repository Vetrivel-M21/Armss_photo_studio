package dto

import (
	"encoding/json"
	"time"
)

type TemplateResponse struct {
	ID                 string          `json:"id"`
	Title              string          `json:"title"`
	ImageURL           string          `json:"imageUrl"`
	ThumbURL           string          `json:"thumbUrl,omitempty"`
	OccasionTags       []string        `json:"occasionTags"`
	Season             string          `json:"season,omitempty"`
	StyleProfile       json.RawMessage `json:"styleProfile,omitempty"`
	StyleProfileStatus string          `json:"styleProfileStatus"` // "ready" | "pending_manual"
	IsActive           bool            `json:"isActive"`
	CreatedAt          time.Time       `json:"createdAt"`
}

// PatchTemplateRequest lets an admin correct the extracted style profile by
// hand (spec §8: "an admin's correction beats a model's guess") and manage
// tags/active state. All fields optional — only what's present is applied.
type PatchTemplateRequest struct {
	StyleProfile json.RawMessage `json:"styleProfile,omitempty"`
	OccasionTags *[]string       `json:"occasionTags,omitempty"`
	Season       *string         `json:"season,omitempty"`
	IsActive     *bool           `json:"isActive,omitempty"`
}
