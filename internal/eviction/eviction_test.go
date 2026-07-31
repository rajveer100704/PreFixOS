package eviction

import (
	"testing"

	"prefixos/internal/interfaces"
)

func TestLRUEvictionPolicy(t *testing.T) {
	lru := NewLRUEvictionPolicy()

	lru.OnInsert(100, 1, 1.0)
	lru.OnInsert(200, 1, 1.0)
	lru.OnInsert(300, 1, 1.0)

	// Access 100 to make it MRU
	lru.OnAccess(100, 1)

	// Victim should be 200 (oldest untouched)
	victim, err := lru.SelectVictim()
	if err != nil {
		t.Fatalf("unexpected error selecting victim: %v", err)
	}
	if victim != 200 {
		t.Fatalf("expected victim 200, got %d", victim)
	}
}

func TestSLRUEvictionPolicy(t *testing.T) {
	slru := NewSLRUEvictionPolicy(2)

	slru.OnInsert(1, 1, 1.0)
	slru.OnInsert(2, 1, 1.0)

	// Promote 1 to protected queue
	slru.OnAccess(1, 1)

	// Victim should come from probation queue (which is 2)
	victim, err := slru.SelectVictim()
	if err != nil {
		t.Fatalf("unexpected error selecting victim: %v", err)
	}
	if victim != 2 {
		t.Fatalf("expected victim 2 from probation, got %d", victim)
	}
}

func TestEvictionPolicyInterfaceCompliance(t *testing.T) {
	policies := []interfaces.EvictionPolicy{
		NewLRUEvictionPolicy(),
		NewSLRUEvictionPolicy(10),
		NewTreeLRU(nil, nil),
	}

	for _, p := range policies {
		if p == nil {
			t.Fatal("expected non-nil eviction policy")
		}
	}
}
