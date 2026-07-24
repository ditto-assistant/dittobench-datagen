package main

import (
	"math"
	"reflect"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/gen"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// TestErrDrawDeterministicAndCalibrated pins that the mid-tier error draw is
// byte-reproducible (same key+salt → same outcome) and that its empirical fail
// rate over many keys tracks the requested rate (it is a hash-uniform draw,
// not a biased one).
func TestErrDrawDeterministicAndCalibrated(t *testing.T) {
	if errDraw("k1", "A", 0.3) != errDraw("k1", "A", 0.3) {
		t.Fatal("errDraw is not deterministic")
	}
	if errDraw("k1", "A", 0) != 1 {
		t.Fatal("rate 0 must always pass")
	}
	if errDraw("k1", "A", 1) != 0 {
		t.Fatal("rate 1 must always fail")
	}
	for _, rate := range []float64{0.05, 0.25} {
		fails := 0
		n := 20000
		for i := 0; i < n; i++ {
			if errDraw(string(rune('a'+i%26))+string(rune('A'+(i/26)%26))+string(rune('0'+(i/676)%10))+string(rune(i)), "A", rate) == 0 {
				fails++
			}
		}
		got := float64(fails) / float64(n)
		if math.Abs(got-rate) > 0.02 {
			t.Errorf("empirical fail rate %.3f for requested %.2f", got, rate)
		}
	}
}

// TestEvalSeedDeterministic pins that scoring one generated dataset twice
// yields identical per-strategy results — the study's noise floor is zero, so
// every reported spread is dataset + modeled-agent variance, not tooling noise.
func TestEvalSeedDeterministic(t *testing.T) {
	prof, _ := gen.ProfileForVersion("full", protocol.BenchVersionV7)
	a, err := gen.GenerateDataset(42, prof, protocol.BenchVersionV7)
	if err != nil {
		t.Fatal(err)
	}
	mk := func() *versionResult {
		vr := &versionResult{Runs: map[string][]seedResult{}, Cases: map[string][][]caseScore{}}
		vr.Seeds = []int64{42}
		evalSeed(vr, a, protocol.BenchVersionV7, 42)
		return vr
	}
	v1, v2 := mk(), mk()
	if !reflect.DeepEqual(v1.Runs, v2.Runs) || !reflect.DeepEqual(v1.StrongB, v2.StrongB) || !reflect.DeepEqual(v1.StrongExp, v2.StrongExp) {
		t.Fatal("evalSeed is not deterministic across identical inputs")
	}
	// The oracle strategy must score a clean 1.0 composite: the study's ceiling
	// rides the same answer-key sufficiency invariant as TestV7OracleSolvable.
	if v1.OracleFailures != 0 {
		t.Fatalf("oracle failed %d cases", v1.OracleFailures)
	}
	or := v1.Runs["oracle"][0]
	if or.Composite != 1 || or.MemMean != 1 || or.ToolMean != 1 {
		t.Fatalf("oracle composite %v", or)
	}
	// The strong tier sits between the naive tiers and the oracle — the
	// decision-boundary regime the study models.
	st := v1.Runs["strong"][0].Composite
	if st <= v1.Runs["overlap"][0].Composite || st >= 1 {
		t.Fatalf("strong composite %.3f is not in the mid tier", st)
	}
	// Expected (structural) composite is within Bernoulli reach of the realized
	// one: |realized - expected| under 10 SDs of the ~240-case binomial.
	if math.Abs(st-v1.StrongExp[0]) > 0.15 {
		t.Fatalf("strong realized %.3f vs expected %.3f", st, v1.StrongExp[0])
	}
}

// TestSeedsForMath pins the confirmation-seed formula against a hand-computed
// value: n = ceil(((z_a+z_b)*sd/gap)^2).
func TestSeedsForMath(t *testing.T) {
	if got := seedsFor(0.0279, 0.01, 1.645, 1.645); got != 85 {
		t.Errorf("seedsFor(0.0279, 0.01, 95%% power) = %d, want 85", got)
	}
	if got := seedsFor(0.0259, 0.005, 1.645, 0); got != 73 {
		t.Errorf("seedsFor(0.0259, 0.005, detect) = %d, want 73", got)
	}
	if got := seedsFor(0.001, 0.5, 1.645, 1.645); got != 1 {
		t.Errorf("tiny sd must floor at 1 seed, got %d", got)
	}
}
