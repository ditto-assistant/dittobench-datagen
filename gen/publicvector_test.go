package gen

import "testing"

// TestKnownVector pins the canonical hash of a fixed seed so any change that
// alters generator output is caught in review. The same value is produced by the
// validators' in-tree generator, so this is also the byte-parity anchor between
// this public repo and the private scoring service. Update it only on a
// deliberate generator change (v2 pre-ship hardening) or a bench_version bump.
func TestKnownVector(t *testing.T) {
	const (
		seed = int64(123456789)
		want = "b56119c38bd887a0b62af7542d921ae22b4b6f268aff969b4eb52e450a4d21d3"
	)
	prof, _ := ProfileFor("full")
	got, _, err := GenerateDataset(seed, prof).SHA256Hex()
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if got != want {
		t.Fatalf("known-vector hash drift for seed %d full:\n got %s\nwant %s", seed, got, want)
	}
}

// TestSameSeedSameBytes is the core determinism guarantee: one seed, one artifact.
func TestSameSeedSameBytes(t *testing.T) {
	prof, _ := ProfileFor("full")
	a, _, err := GenerateDataset(42, prof).SHA256Hex()
	if err != nil {
		t.Fatalf("hash a: %v", err)
	}
	b, _, err := GenerateDataset(42, prof).SHA256Hex()
	if err != nil {
		t.Fatalf("hash b: %v", err)
	}
	if a != b {
		t.Fatalf("same seed produced different bytes: %s vs %s", a, b)
	}
}
