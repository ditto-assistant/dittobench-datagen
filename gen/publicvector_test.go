package gen

import "testing"

// TestKnownVector pins the canonical hash of a fixed seed so any change that
// alters generator output is caught in review. The validators run this same
// module, so this is also the byte-parity anchor between this public repo and the
// private scoring service. The hash is a function of (seed, bench_version): a
// bench_version bump rotates the surface (RotateSeed) and moves it by design.
// Update it only on a deliberate generator change or a bench_version bump.
func TestKnownVector(t *testing.T) {
	const (
		seed = int64(123456789)
		// Deliberate change (judge-free grading): MemoryCase gained
		// answer_kind/answer_items/distractor_answers and the binary temporal
		// leftover question was dropped, both of which move the artifact bytes.
		want = "73781f0f59c1e5bc0864463ff3e2f658d6f3d525923324d2add73b88351f323d"
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
