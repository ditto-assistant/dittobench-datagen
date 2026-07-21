package gen

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/grade"
	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// TestTyposNeverTouchCoinedOrProtected is the answer-integrity guard: the informal
// noise must never mutate a coined token (every graded value/distractor/sentinel)
// or an explicitly protected substring (the multi-hop join name), or cases become
// ungradeable / unwinnable. Every CoinShaped surface shape is exercised.
func TestTyposNeverTouchCoinedOrProtected(t *testing.T) {
	coined := []string{}
	for _, s := range []int64{1, 7, 42, 99, 123456789, 31337, 2, 5, 8, 13} {
		coined = append(coined, persona.CoinShaped(s, "a"), persona.CoinShaped(s, "b|2"))
	}
	name := "Gunnar"
	for _, tok := range coined {
		base := "oh by the way " + name + " finally named their sailboat " + tok + " last weekend"
		for seed := int64(0); seed < 300; seed++ {
			r := rand.New(rand.NewSource(seed))
			out := applyTypos(r, base, []string{name, tok})
			if !strings.Contains(out, tok) {
				t.Fatalf("coined token %q was corrupted: %q", tok, out)
			}
			if !strings.Contains(out, name) {
				t.Fatalf("protected name %q was corrupted: %q", name, out)
			}
		}
	}
}

// TestTyposDeterministicAndActive: same rng stream => same output (reproducibility),
// and across seeds the noise actually fires (not a silent no-op) while sometimes
// leaving a string untouched (the count is a real 0..bound draw).
func TestTyposDeterministicAndActive(t *testing.T) {
	base := "quick note my best friend just moved to a little town called somewhere nice"
	// determinism
	a := applyTypos(rand.New(rand.NewSource(12345)), base, nil)
	b := applyTypos(rand.New(rand.NewSource(12345)), base, nil)
	if a != b {
		t.Fatalf("applyTypos not deterministic:\n %q\n %q", a, b)
	}
	changed, unchanged := 0, 0
	for seed := int64(0); seed < 200; seed++ {
		out := applyTypos(rand.New(rand.NewSource(seed)), base, nil)
		if out == base {
			unchanged++
		} else {
			changed++
		}
	}
	if changed == 0 {
		t.Fatal("typos never fired — informal noise is inert")
	}
	if unchanged == 0 {
		t.Fatal("typos always fired — the 0-typo draw never happens (count not bounded/variable)")
	}
}

// TestV6TyposPreserveWinnableAnswers: under v6 informal noise, the coined answer of
// every value-recall family is still seeded VERBATIM in the haystack, so the case
// stays winnable — the topic is noisy, the answer is not.
func TestV6TyposPreserveWinnableAnswers(t *testing.T) {
	valueFamily := map[string]bool{
		QTMultiHop: true, QTTemporalDepth: true,
		QTDeclarativeRead: true, QTDeclarativeBehavior: true,
	}
	for _, seed := range []int64{31337, 99, 7, 424242} {
		r, _ := NewRNGForVersion(seed, protocol.BenchVersionV6)
		s, err := GenerateMemorySuiteForVersion(r, seed, 90, 4, 0.2, protocol.BenchVersionV6)
		if err != nil {
			t.Fatal(err)
		}
		// The corpus a harness sees: passive haystack pairs PLUS the /run case turns
		// (a declarative preference's value is stated in its earlier write-turn case,
		// not in a haystack pair).
		var hay strings.Builder
		for _, w := range s.Waves {
			for _, p := range w.Pairs {
				hay.WriteString(p.Prompt + " " + p.Response + " ")
			}
		}
		for _, sc := range s.Cases {
			hay.WriteString(sc.Case.Question + " ")
		}
		haystack := hay.String()
		n := 0
		for _, sc := range s.Cases {
			if !valueFamily[sc.Case.QuestionType] {
				continue
			}
			n++
			if !grade.Hit(sc.Case.ExpectedAnswer, haystack) {
				t.Errorf("seed %d: %s answer %q not seeded verbatim under v6 typos — unwinnable",
					seed, sc.Case.QuestionType, sc.Case.ExpectedAnswer)
			}
		}
		if n == 0 {
			t.Fatalf("seed %d: no value-family cases generated", seed)
		}
	}
}
