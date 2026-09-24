package imagegen

import (
	"fmt"
	"sync"
)

// Registry maps provider name to implementation. Get never falls back to a
// different provider on failure — cost and style both change between vendors
// (spec §7), so a failure must be visible, not silently substituted.
//
// Guarded by a mutex and swappable at runtime (Swap) — an admin pasting a
// replacement API key for a provider takes effect immediately for every
// in-flight and future call, without restarting the server.
type Registry struct {
	mu          sync.RWMutex
	providers   map[string]ImageProvider
	defaultName string
}

func NewRegistry(defaultName string, providers ...ImageProvider) *Registry {
	m := make(map[string]ImageProvider, len(providers))
	for _, p := range providers {
		m[p.Name()] = p
	}
	return &Registry{providers: m, defaultName: defaultName}
}

// Get returns the named provider, or the server default if name is empty.
func (r *Registry) Get(name string) (ImageProvider, error) {
	if name == "" {
		name = r.defaultName
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown image provider %q", name)
	}
	return p, nil
}

// Swap atomically replaces the named provider (e.g. rebuilt with a new API
// key) — every subsequent Get(name) call, from any goroutine, returns it
// immediately.
func (r *Registry) Swap(name string, p ImageProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = p
}

// SetDefault changes the default image provider name at runtime.
func (r *Registry) SetDefault(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultName = name
}

// DefaultName returns the current default image provider name.
func (r *Registry) DefaultName() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultName
}

// List returns all registered provider names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	return names
}
