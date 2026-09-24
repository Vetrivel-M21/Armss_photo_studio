// Package pipeline orchestrates the agent pipeline: the project state machine,
// stage re-run/supersede logic, and the SSE event bus that lets the UI show
// each agent completing instead of one long spinner (spec §9).
package pipeline

import "sync"

type EventStatus string

const (
	StatusRunning EventStatus = "running"
	StatusDone    EventStatus = "done"
	StatusError   EventStatus = "error"
)

// Event matches the SSE wire shape from spec §9 exactly.
type Event struct {
	Stage     string      `json:"stage"`
	Status    EventStatus `json:"status"`
	Payload   any         `json:"payload,omitempty"`
	TokensIn  int         `json:"tokensIn"`
	TokensOut int         `json:"tokensOut"`
}

// Bus is an in-process, per-project pub/sub used to stream pipeline progress.
// Single-process and in-memory is a known scaling limit (would need e.g. Redis
// pub/sub for multi-instance deployment) — not built now, since this app runs
// as one process (no speculative abstraction for a scale we're not at).
type Bus struct {
	mu   sync.Mutex
	subs map[string][]chan Event
}

func NewBus() *Bus {
	return &Bus{subs: make(map[string][]chan Event)}
}

// Subscribe returns a channel of events for projectID and an unsubscribe func
// that must be called (typically via defer) when the caller stops listening.
func (b *Bus) Subscribe(projectID string) (<-chan Event, func()) {
	ch := make(chan Event, 16)
	b.mu.Lock()
	b.subs[projectID] = append(b.subs[projectID], ch)
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subs[projectID]
		for i, c := range subs {
			if c == ch {
				b.subs[projectID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}
	return ch, unsubscribe
}

// Publish sends evt to every current subscriber of projectID. Slow subscribers
// have events dropped rather than blocking the publisher (agent pipeline).
func (b *Bus) Publish(projectID string, evt Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subs[projectID] {
		select {
		case ch <- evt:
		default:
		}
	}
}
