package gen

import (
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

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
	artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV2)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got, _, err := artifact.SHA256Hex()
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if got != want {
		t.Fatalf("known-vector hash drift for seed %d full:\n got %s\nwant %s", seed, got, want)
	}
}

// TestV3KnownVector publishes the first rotated benchmark contract alongside
// the immutable v2 vector above. Both remain pinned forever.
func TestV3KnownVector(t *testing.T) {
	const (
		seed = int64(123456789)
		want = "cdb0e6431b47a98492059e199dc8bd2567be00d1d299d55cdca0a67a26abb32a"
	)
	prof, _ := ProfileFor("full")
	artifact, err := GenerateDataset(seed, prof, protocol.BenchVersionV3)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got, _, err := artifact.SHA256Hex()
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if got != want {
		t.Fatalf("v3 known-vector hash drift for seed %d full:\n got %s\nwant %s", seed, got, want)
	}
}

func TestVersionChangesCanonicalBytes(t *testing.T) {
	prof, _ := ProfileFor("full")
	v2, err := GenerateDataset(42, prof, protocol.BenchVersionV2)
	if err != nil {
		t.Fatal(err)
	}
	v3, err := GenerateDataset(42, prof, protocol.BenchVersionV3)
	if err != nil {
		t.Fatal(err)
	}
	v2Hash, _, _ := v2.SHA256Hex()
	v3Hash, _, _ := v3.SHA256Hex()
	if v2Hash == v3Hash {
		t.Fatal("v2 and v3 produced identical canonical bytes")
	}
	if v2.BenchVersion != 2 || v3.BenchVersion != 3 || v2.GeneratedAt == v3.GeneratedAt {
		t.Fatalf("version provenance not rotated: v2=%+v v3=%+v", v2, v3)
	}
}

func TestUnsupportedVersionRejected(t *testing.T) {
	prof, _ := ProfileFor("small")
	if _, err := GenerateDataset(42, prof, 4); err == nil {
		t.Fatal("unsupported version accepted")
	}
}

// TestSameSeedSameBytes is the core determinism guarantee: one seed, one artifact.
func TestSameSeedSameBytes(t *testing.T) {
	prof, _ := ProfileFor("full")
	for _, version := range []int{protocol.BenchVersionV2, protocol.BenchVersionV3} {
		artifactA, err := GenerateDataset(42, prof, version)
		if err != nil {
			t.Fatalf("v%d generate a: %v", version, err)
		}
		a, _, err := artifactA.SHA256Hex()
		if err != nil {
			t.Fatalf("v%d hash a: %v", version, err)
		}
		artifactB, err := GenerateDataset(42, prof, version)
		if err != nil {
			t.Fatalf("v%d generate b: %v", version, err)
		}
		b, _, err := artifactB.SHA256Hex()
		if err != nil {
			t.Fatalf("v%d hash b: %v", version, err)
		}
		if a != b {
			t.Fatalf("v%d same seed produced different bytes: %s vs %s", version, a, b)
		}
	}
}
