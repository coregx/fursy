package middleware

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// TDD: insertOrd slice grows unbounded — Cleanup removes from map but not
// from insertOrd. After many cycles of add→cleanup, insertOrd has thousands
// of stale entries while map is empty.
func TestRateLimit_InsertOrdMemoryLeak(t *testing.T) {
	store := newInMemoryStore(100000) // high max so eviction doesn't trigger

	// Phase 1: Add 1000 keys.
	for i := 0; i < 1000; i++ {
		store.GetLimiter(fmt.Sprintf("k%d", i), rate.Limit(10), 20)
	}

	// Phase 2: Expire all keys and cleanup.
	store.mu.Lock()
	for k, entry := range store.limiters {
		entry.lastAccess = time.Now().Add(-2 * time.Hour)
		store.limiters[k] = entry
	}
	store.mu.Unlock()
	store.Cleanup(1 * time.Hour)

	// Phase 3: Verify map is empty.
	store.mu.Lock()
	mapLen := len(store.limiters)
	store.mu.Unlock()

	if mapLen != 0 {
		t.Fatalf("expected 0 limiters after cleanup, got %d", mapLen)
	}

	// Phase 4: LRU list should also be cleaned up, not 1000 stale entries.
	store.mu.Lock()
	lruLen := store.lruList.Len()
	indexLen := len(store.lruIndex)
	store.mu.Unlock()

	if lruLen > 100 {
		t.Errorf("REGRESSION: lruList has %d stale entries after cleanup (should be 0)", lruLen)
	}
	if indexLen > 100 {
		t.Errorf("REGRESSION: lruIndex has %d stale entries after cleanup (should be 0)", indexLen)
	}
}

// TDD: Re-added key after cleanup creates duplicate in insertOrd.
// When eviction happens, the stale entry deletes the fresh one.
func TestRateLimit_DuplicateInsertOrd(t *testing.T) {
	store := newInMemoryStore(3)

	// Add key "A".
	store.GetLimiter("A", rate.Limit(10), 20)
	store.GetLimiter("B", rate.Limit(10), 20)

	// Expire "A" and cleanup.
	store.mu.Lock()
	if e, ok := store.limiters["A"]; ok {
		e.lastAccess = time.Now().Add(-2 * time.Hour)
	}
	store.mu.Unlock()
	store.Cleanup(1 * time.Hour)

	// Re-add "A" — now "A" is in insertOrd twice (stale + fresh).
	store.GetLimiter("A", rate.Limit(10), 20)
	store.GetLimiter("C", rate.Limit(10), 20)

	// Add "D" — triggers eviction. Should evict "B" (oldest active), not "A".
	store.GetLimiter("D", rate.Limit(10), 20)

	store.mu.Lock()
	_, aExists := store.limiters["A"]
	store.mu.Unlock()

	if !aExists {
		t.Error("REGRESSION: re-added key 'A' was evicted due to stale insertOrd duplicate")
	}
}
