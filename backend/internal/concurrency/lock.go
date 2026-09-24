package concurrency

import (
	"errors"
	"fmt"
	"sync"
)

// ErrAlreadyInProgress indicates a duplicate or concurrent action was attempted
// while an existing one is currently running.
var ErrAlreadyInProgress = errors.New("an operation is already in progress for this project")

// Registry manages thread-safe, in-flight operation locks across goroutines
// to prevent duplicate AI generations, parallel pipeline runs, or redundant billing.
type Registry struct {
	mu       sync.Mutex
	inFlight map[string]struct{}
}

// NewRegistry creates a fresh concurrency in-flight lock registry.
func NewRegistry() *Registry {
	return &Registry{
		inFlight: make(map[string]struct{}),
	}
}

// TryLock attempts to acquire an exclusive lock for the given project and action.
// If already in-flight, it returns nil and false.
// If successfully acquired, it returns an idempotent unlock function and true.
func (r *Registry) TryLock(projectID, action string) (func(), bool) {
	key := fmt.Sprintf("%s:%s", projectID, action)
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.inFlight[key]; exists {
		return nil, false
	}

	r.inFlight[key] = struct{}{}

	var once sync.Once
	unlock := func() {
		once.Do(func() {
			r.mu.Lock()
			delete(r.inFlight, key)
			r.mu.Unlock()
		})
	}
	return unlock, true
}

// IsInFlight reports whether an action on a project is currently being executed.
func (r *Registry) IsInFlight(projectID, action string) bool {
	key := fmt.Sprintf("%s:%s", projectID, action)
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.inFlight[key]
	return exists
}
