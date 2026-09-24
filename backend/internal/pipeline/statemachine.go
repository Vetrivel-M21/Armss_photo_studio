package pipeline

import "armss-poster-studio/internal/store/mysql"

// States is the project state machine from spec §4, in forward order. Project
// state moves monotonically forward; edits to an earlier stage don't regress
// it — instead, downstream artifact "staleness" is computed at read time from
// stage_outputs.superseded and renders.created_at (see stage_outputs_repo.go
// and renders_repo.go). "approved" is unused until the Reviewer agent ships
// (out of scope for this build) but stays in the enum so the schema doesn't
// need to change later.
var States = []string{
	mysql.StateDraft,
	mysql.StateBriefing,
	mysql.StateConceptsReady,
	mysql.StateDeveloping,
	mysql.StateRenderReady,
	mysql.StateComposited,
	mysql.StateApproved,
}

func stateIndex(state string) int {
	for i, s := range States {
		if s == state {
			return i
		}
	}
	return -1
}

// IsAtLeast reports whether current has reached or passed target in the
// forward ordering.
func IsAtLeast(current, target string) bool {
	return stateIndex(current) >= stateIndex(target)
}
