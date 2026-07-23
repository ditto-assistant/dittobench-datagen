package gen

import (
	"sort"
	"testing"
	"time"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// approxTokens is a chars/4 estimate. It is deliberately crude: these tests
// assert an order of magnitude, and a tokenizer dependency would buy precision
// nobody needs here.
func approxTokens(pairs []protocol.MemoryPair) int {
	n := 0
	for _, p := range pairs {
		n += len(p.Prompt) + len(p.Response)
	}
	return n / 4
}

func primaryPairs(t *testing.T, seed int64, version int) []protocol.MemoryPair {
	t.Helper()
	prof, _ := ProfileForVersion("full", version)
	art, err := GenerateDataset(seed, prof, version)
	if err != nil {
		t.Fatalf("v%d generate: %v", version, err)
	}
	var out []protocol.MemoryPair
	for _, w := range art.MemoryWaves {
		if w.UserID != "" && w.UserID != PrimaryUser {
			continue
		}
		out = append(out, w.Pairs...)
	}
	return out
}

// longMemEvalSmallTokens is the haystack size LongMemEval_S gives a question
// (arXiv:2410.10813). It is the parity target v8 exists to hit; the floor below
// sits under it so ordinary per-seed variation does not fail CI.
const longMemEvalSmallTokens = 115_000

// TestV8HaystackReachesLongMemEvalScale is the reason bench_version 8 exists.
//
// Through v7 the scored history was ~4.5k tokens: a harness could hold all of it
// in context and never retrieve, so the memory score measured reading rather than
// memory. This pins the floor across seeds so a later change cannot quietly shrink
// the history back to a size that makes retrieval optional.
func TestV8HaystackReachesLongMemEvalScale(t *testing.T) {
	const floorTokens = 95_000
	for _, seed := range []int64{123456789, 42, 987654321, 555} {
		pairs := primaryPairs(t, seed, protocol.BenchVersionV8)
		tokens := approxTokens(pairs)
		if tokens < floorTokens {
			t.Errorf("seed %d: primary haystack is ~%d tokens, want at least %d (LongMemEval_S is ~%d)",
				seed, tokens, floorTokens, longMemEvalSmallTokens)
		}

		sessions := map[string]bool{}
		var first, last time.Time
		for _, p := range pairs {
			sessions[p.SessionID] = true
			ts, err := time.Parse(time.RFC3339, p.Timestamp)
			if err != nil {
				t.Fatalf("seed %d: bad timestamp %q: %v", seed, p.Timestamp, err)
			}
			if first.IsZero() || ts.Before(first) {
				first = ts
			}
			if ts.After(last) {
				last = ts
			}
		}
		if len(sessions) < 40 {
			t.Errorf("seed %d: %d sessions, want at least 40 (LongMemEval_S pairs a question with 30-50)",
				seed, len(sessions))
		}
		if span := last.Sub(first); span < 365*24*time.Hour {
			t.Errorf("seed %d: history spans %.0f days, want at least a year so elapsed-duration "+
				"and point-in-time answers range over a real horizon", seed, span.Hours()/24)
		}
	}
}

// TestV8IsMateriallyDeeperThanV7 pins the contrast the release is defined by, so
// "v8 is the deep-history contract" stays a fact about the bytes rather than a
// claim in a changelog.
func TestV8IsMateriallyDeeperThanV7(t *testing.T) {
	const seed = int64(123456789)
	v7 := approxTokens(primaryPairs(t, seed, protocol.BenchVersionV7))
	v8 := approxTokens(primaryPairs(t, seed, protocol.BenchVersionV8))
	if v8 < 10*v7 {
		t.Fatalf("v8 haystack is ~%d tokens vs v7's ~%d; the deep-history contract should be "+
			"an order of magnitude deeper", v8, v7)
	}
}

// TestV8GradedCaseCountStaysFlat is the COST guard.
//
// The deep-history release scales what a harness must retrieve from, not how many
// questions it must answer: ingest happens once per submission, while every graded
// case is an inference round trip. Letting the case count drift up with the
// haystack would turn a retrieval benchmark into a bill. This fails if a future
// change scales cases alongside history.
func TestV8GradedCaseCountStaysFlat(t *testing.T) {
	v7, _ := ProfileForVersion("full", protocol.BenchVersionV7)
	v8, _ := ProfileForVersion("full", protocol.BenchVersionV8)
	if v8.Mem > v7.Mem*5/4 {
		t.Errorf("v8 memory cases %d vs v7 %d: graded volume should stay near-flat while "+
			"history scales", v8.Mem, v7.Mem)
	}
	if v8.Tools > v7.Tools {
		t.Errorf("v8 tool cases %d vs v7 %d: tool volume should not grow with history depth",
			v8.Tools, v7.Tools)
	}
}

// TestV8FactBeatsAreNotSeparableByLength closes the shortcut that a large filler
// layer would otherwise open.
//
// If evidence-bearing turns were one-liners in a haystack of paragraphs, the
// cheapest winning strategy would be "index the short turns and ignore the rest",
// which beats retrieval without doing any. v8 elaborates fact beats from the same
// clause pools the background threads use; this asserts the two populations really
// do overlap, by checking that the median lengths are within a small factor and
// that the shortest decile is not overwhelmingly evidence.
func TestV8FactBeatsAreNotSeparableByLength(t *testing.T) {
	const seed = int64(123456789)
	plan, err := persona.BuildPlanForVersion(seed, personaOptsForVersion(100, protocol.BenchVersionV8), protocol.BenchVersionV8)
	if err != nil {
		t.Fatal(err)
	}
	pairs, evidence := RenderHaystack(plan)
	isEvidence := map[string]bool{}
	for _, pid := range evidence {
		isEvidence[pid] = true
	}

	var factLens, fillerLens []int
	type sized struct {
		n    int
		fact bool
	}
	var all []sized
	for _, p := range pairs {
		n := len(p.Prompt) + len(p.Response)
		all = append(all, sized{n, isEvidence[p.PairID]})
		if isEvidence[p.PairID] {
			factLens = append(factLens, n)
		} else {
			fillerLens = append(fillerLens, n)
		}
	}
	if len(factLens) == 0 || len(fillerLens) == 0 {
		t.Fatalf("degenerate haystack: %d fact pairs, %d filler pairs", len(factLens), len(fillerLens))
	}
	median := func(xs []int) int {
		sort.Ints(xs)
		return xs[len(xs)/2]
	}
	mf, mn := median(factLens), median(fillerLens)
	ratio := float64(mf) / float64(mn)
	if ratio < 0.5 || ratio > 2.0 {
		t.Errorf("median fact-turn length %d vs filler %d (ratio %.2f): the two populations must "+
			"overlap, or length alone separates evidence from noise", mf, mn, ratio)
	}

	// The shortest tenth of the haystack must not be mostly evidence, or
	// "read the short turns first" is a cheap retrieval prior.
	sort.Slice(all, func(i, j int) bool { return all[i].n < all[j].n })
	decile := len(all) / 10
	hits := 0
	for _, s := range all[:decile] {
		if s.fact {
			hits++
		}
	}
	base := float64(len(factLens)) / float64(len(all))
	if got := float64(hits) / float64(decile); got > base*3 {
		t.Errorf("%.0f%% of the shortest decile is evidence vs a %.0f%% base rate: short turns "+
			"leak where the answers are", got*100, base*100)
	}
}
