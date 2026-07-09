package main

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// A benchmark whose entire spread comes from which seed you drew (both
// categories move together, run to run) must be flagged seed-dominated.
func TestSeedDominated(t *testing.T) {
	runs := []run{
		{Seed: 1, PerCase: []perCase{{"a", 0}, {"b", 0}}},
		{Seed: 2, PerCase: []perCase{{"a", 1}, {"b", 1}}},
	}
	rep := analyze(runs)
	if rep.DominantFacet != "seed" {
		t.Fatalf("dominant facet = %q, want seed", rep.DominantFacet)
	}
	if rep.Variance.Item != 0 {
		t.Fatalf("item component = %v, want 0 (categories never differ within a seed)", rep.Variance.Item)
	}
}

// A benchmark whose spread comes from which category you're in (easy always
// passes, hard always fails, regardless of seed) must be flagged item-dominated.
func TestItemDominated(t *testing.T) {
	runs := []run{
		{Seed: 1, PerCase: []perCase{{"easy", 1}, {"hard", 0}}},
		{Seed: 2, PerCase: []perCase{{"easy", 1}, {"hard", 0}}},
	}
	rep := analyze(runs)
	if rep.DominantFacet != "item" {
		t.Fatalf("dominant facet = %q, want item", rep.DominantFacet)
	}
	if rep.Variance.Seed != 0 {
		t.Fatalf("seed component = %v, want 0 (seeds never differ)", rep.Variance.Seed)
	}
}

// Saturated (everyone passes) and floor (everyone fails) categories carry no
// discrimination and must be flagged so they can be retired.
func TestSaturationFlags(t *testing.T) {
	runs := []run{
		{Seed: 1, PerCase: []perCase{{"trivial", 1}, {"impossible", 0}, {"mixed", 1}}},
		{Seed: 2, PerCase: []perCase{{"trivial", 1}, {"impossible", 0}, {"mixed", 0}}},
	}
	rep := analyze(runs)
	flags := map[string]string{}
	for _, c := range rep.Categories {
		flags[c.Category] = c.Flag
	}
	if flags["trivial"] != "saturated" {
		t.Errorf("trivial flag = %q, want saturated", flags["trivial"])
	}
	if flags["impossible"] != "floor" {
		t.Errorf("impossible flag = %q, want floor", flags["impossible"])
	}
	if flags["mixed"] != "" {
		t.Errorf("mixed flag = %q, want empty", flags["mixed"])
	}
}

func TestBetweenGroupZeroWhenNoSpread(t *testing.T) {
	groups := [][]float64{{0.5, 0.5}, {0.5, 0.5}}
	if got := betweenGroup(groups, 0.5); !approx(got, 0) {
		t.Fatalf("betweenGroup = %v, want 0", got)
	}
}
