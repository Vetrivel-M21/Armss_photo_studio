package dto

import (
	"encoding/json"
	"time"
)

type BrandAssetResponse struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	ImageURL  string          `json:"imageUrl"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Version   int             `json:"version"`
	IsActive  bool            `json:"isActive"`
	CreatedAt time.Time       `json:"createdAt"`
}
