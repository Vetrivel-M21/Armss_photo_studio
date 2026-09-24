package pipeline

import "armss-poster-studio/internal/agents"

// ConceptsPayload wraps the Creative Director's three concepts for storage as
// a single stage_outputs row (stage="concepts").
type ConceptsPayload struct {
	Concepts []agents.Concept `json:"concepts"`
}

// SelectionPayload records which concept the user picked, stored as its own
// stage_outputs row (stage="concept_selection") — an addition beyond the
// spec's literal stage enum, keeping the choice in the same auditable,
// superseded-on-rerun history as everything else.
type SelectionPayload struct {
	SelectedIndex int            `json:"selectedIndex"`
	Concept       agents.Concept `json:"concept"`
}
