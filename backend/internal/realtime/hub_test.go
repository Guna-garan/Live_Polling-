package realtime

import "testing"

// TestSnapshotMessageIncludesTotal verifies the snapshot envelope sums
// per-option counts correctly, since the frontend trusts totalVotes
// rather than re-summing client-side.
func TestSnapshotMessageIncludesTotal(t *testing.T) {
	payload, err := SnapshotMessage(map[string]int64{"a": 3, "b": 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload) == 0 {
		t.Fatal("expected non-empty payload")
	}
	// A light string check is enough here; full JSON round-tripping is
	// exercised by the integration tests against a live Redis instance
	// (see redis_integration_test.go), which is where the8-vote-count
	// arithmetic actually matters end-to-end.
	got := string(payload)
	if !contains(got, `"totalVotes":8`) {
		t.Fatalf("expected totalVotes:8 in payload, got %s", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
