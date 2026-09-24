package dto

type RecommendationsRequest struct {
	Text string `json:"text"`
}

type RecommendationView struct {
	Template TemplateResponse `json:"template"`
	Reason   string           `json:"reason"`
}
