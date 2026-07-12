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
		//
		// Moved again (anti-gaming addendum N2, bench_version 2): the metamorphic
		// invariance twin became a j=3 sibling FAMILY (invarianceTwins) instead of
		// a seed-selected pair, so the run emits one more consistency case and the
		// twin phrasing-set selection changed.
		//
		// Moved again (prod hardening P1+P6, bench_version 2): write-then-read
		// lifecycle chains joined the memory suite (gen/lifecycle.go) and the
		// point-in-time modality joined question derivation, changing the case
		// mix, wave-0 pairs, and RNG draw counts.
		//
		// Moved again (metamorphic multi-family, bench_version 2): question
		// derivation now emits several invariance families (invarianceTwins) and
		// the memory suite selects twinFamiliesFor(n) of them (full=3), so the twin
		// case count and attribute selection changed. This smooths the
		// metamorphic-consistency composite factor from a per-run coin flip.
		want = "dfb4fc243d7d3e84bb4e896d5873bbc9bda114e16f5215f913c13adbfbc4a7fe"
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
