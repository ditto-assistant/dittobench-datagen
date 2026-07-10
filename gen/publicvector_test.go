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
		// Deliberate change (audit follow-up item A, bench_version 2): the five
		// audited low-variety tool categories moved from flat template lists to
		// CFG grammar expansion (persona.Expand), which changes both the emitted
		// prompts and the per-case RNG draw counts.
		want = "e96291e2dfe1cfcaa42336bd07d63737e6b10e4201fae79cc641adac4f45c5ac"
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
