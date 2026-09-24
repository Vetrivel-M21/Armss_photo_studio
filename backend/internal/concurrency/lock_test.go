package concurrency_test

import (
	"sync"
	"testing"

	"armss-poster-studio/internal/concurrency"
)

func TestRegistry_TryLock(t *testing.T) {
	reg := concurrency.NewRegistry()

	// Initial lock should succeed
	unlock, ok := reg.TryLock("proj-123", "render")
	if !ok || unlock == nil {
		t.Fatalf("expected initial lock to succeed, got ok=%v", ok)
	}

	if !reg.IsInFlight("proj-123", "render") {
		t.Errorf("expected IsInFlight to be true while held")
	}

	// Immediate duplicate lock attempt should fail
	_, duplicateOk := reg.TryLock("proj-123", "render")
	if duplicateOk {
		t.Errorf("expected duplicate lock to be rejected, but it succeeded")
	}

	// Different project or different action should succeed
	unlock2, ok2 := reg.TryLock("proj-456", "render")
	if !ok2 {
		t.Errorf("expected lock for different project to succeed")
	}
	defer unlock2()

	unlock3, ok3 := reg.TryLock("proj-123", "composite")
	if !ok3 {
		t.Errorf("expected lock for different action on same project to succeed")
	}
	defer unlock3()

	// Unlock first lock
	unlock()

	if reg.IsInFlight("proj-123", "render") {
		t.Errorf("expected IsInFlight to be false after unlock")
	}

	// Subsequent lock should now succeed
	unlockAfter, okAfter := reg.TryLock("proj-123", "render")
	if !okAfter {
		t.Errorf("expected re-acquiring lock after unlock to succeed")
	}
	unlockAfter()
}

func TestRegistry_ConcurrentContention(t *testing.T) {
	reg := concurrency.NewRegistry()
	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			unlock, ok := reg.TryLock("shared-project", "generate")
			if ok {
				mu.Lock()
				successCount++
				mu.Unlock()
				// Hold briefly then unlock
				unlock()
			}
		}()
	}

	wg.Wait()

	if successCount == 0 {
		t.Errorf("expected at least 1 successful acquisition under contention")
	}
}
