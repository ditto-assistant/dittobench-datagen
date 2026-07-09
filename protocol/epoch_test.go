package protocol

import "testing"

// RotateSeed must be a deterministic pure function so any auditor recomputes the
// same generation stream from (seed, bench_version).
func TestRotateSeedDeterministic(t *testing.T) {
	for _, s := range []int64{0, 1, -1, 123456789, 1 << 40} {
		if RotateSeed(s) != RotateSeed(s) {
			t.Fatalf("RotateSeed(%d) is not deterministic", s)
		}
	}
}

// The rotation must actually perturb the seed (otherwise it is a no-op and the
// per-version treadmill does nothing), and distinct seeds must stay distinct.
func TestRotateSeedActiveAndInjective(t *testing.T) {
	if RotateSeed(123456789) == 123456789 {
		t.Fatal("RotateSeed is the identity; the version rotation is inert")
	}
	seen := map[int64]int64{}
	for _, s := range []int64{0, 1, 2, 3, 100, 123456789, -42, 1 << 50} {
		r := RotateSeed(s)
		if prev, ok := seen[r]; ok {
			t.Fatalf("RotateSeed collided: %d and %d both map to %d", prev, s, r)
		}
		seen[r] = s
	}
}
