package gen

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// bench_version 7 temporal arithmetic. The v6 non-verbatim cases
// (gen/nonverbatim.go) convert ONE stored quantity between units; the fact and
// the conversion table are both in a single pair. A temporal-arithmetic case
// states a BASE quantity in one session and a CHANGE to it in a later session,
// and asks for the combined result:
//
//	session A: "My rent is $1,850 a month."
//	session B: "The landlord raised my rent by $240 starting next month."
//	question:  "Counting that increase, what am I paying for rent now?" -> 2,090
//
// The answer token appears in NO seeded pair — echoing either stored number
// fails containment — and no single pair even contains the inputs, so the case
// requires cross-session retrieval AND arithmetic. Graded deterministically
// through the accept-set (comma and plain digit forms).
//
// Deterministic per (seed, draw position); v7-gated so v6's bytes are
// untouched.
const QTTempCalc = "temporal-arithmetic"

type tempCalcSuite struct {
	Cases []StagedCase
	Pairs []protocol.MemoryPair
}

func tempCalcEnabled(benchVersion int) bool {
	return benchVersion >= protocol.BenchVersionV7
}

// tempCalcCasesFor is the seed-independent case quota.
func tempCalcCasesFor(n, nWaves int) int {
	switch {
	case nWaves < 2:
		return 0
	case n >= 100:
		return 3
	case n >= 40:
		return 2
	case n >= 20:
		return 1
	default:
		return 0
	}
}

// tcDomain is one arithmetic domain: how the base and the change are stated
// (fmt: the formatted number), how the combined value is asked, and how the
// two numbers combine. Base/delta ranges keep every value positive, the
// result distinct from both inputs, and the phrasing natural.
type tcDomain struct {
	base    []string // fmt: base number
	change  []string // fmt: delta number
	ask     []string
	combine func(base, delta int) int
	baseLo  int
	baseN   int // base = baseLo + step*Intn(baseN)
	step    int
	deltaLo int
	deltaN  int // delta = deltaLo + dstep*Intn(deltaN)
	dstep   int
}

var tcDomains = []tcDomain{
	{ // rent raise: base + delta
		base: []string{
			"My rent is $%s a month.",
			"For the record, I pay $%s a month in rent.",
		},
		change: []string{
			"Heads up — the landlord raised my rent by $%s starting next month.",
			"Ugh, my rent is going up by $%s next month.",
		},
		ask: []string{
			"Counting that increase, what am I paying for rent per month now?",
			"With the raise included, what's my total monthly rent?",
			"What does my rent come to per month after the increase?",
		},
		combine: func(b, d int) int { return b + d },
		baseLo:  1400, baseN: 13, step: 50,
		deltaLo: 60, deltaN: 10, dstep: 20,
	},
	{ // pages read: base + delta
		base: []string{
			"I'm %s pages into the big novel.",
			"Made it %s pages into the novel so far.",
		},
		change: []string{
			"Read another %s pages of the novel tonight.",
			"Got through %s more pages of the novel on the train.",
		},
		ask: []string{
			"How many pages of the novel have I read in total?",
			"What's my total page count on the novel now?",
			"All together, how many pages of the novel am I through?",
		},
		combine: func(b, d int) int { return b + d },
		baseLo:  120, baseN: 40, step: 5,
		deltaLo: 40, deltaN: 24, dstep: 5,
	},
	{ // savings goal: base - delta
		base: []string{
			"I'm saving up $%s for the rafting trip.",
			"The rafting trip I want costs $%s, saving toward it.",
		},
		change: []string{
			"I've put aside $%s for the rafting trip so far.",
			"So far I've managed to set aside $%s for the rafting trip.",
		},
		ask: []string{
			"How much more do I need to save for the rafting trip?",
			"What's left to save for the rafting trip?",
			"How far am I from the rafting trip amount, in dollars?",
		},
		combine: func(b, d int) int { return b - d },
		baseLo:  2200, baseN: 16, step: 100,
		deltaLo: 300, deltaN: 18, dstep: 50,
	},
	{ // PTO days: base - delta
		base: []string{
			"I get %s days of PTO this year.",
			"My PTO allowance is %s days this year.",
		},
		change: []string{
			"I've already used %s of my PTO days.",
			"So far I've burned %s PTO days.",
		},
		ask: []string{
			"How many PTO days do I have left this year?",
			"What's my remaining PTO day count?",
			"After what I've used, how many PTO days remain?",
		},
		combine: func(b, d int) int { return b - d },
		baseLo:  21, baseN: 8, step: 1,
		deltaLo: 4, deltaN: 9, dstep: 1,
	},
}

// tcComma renders n with thousands separators; tcPlain is the bare digit form.
// Both go in the accept-set so "2,090" and "2090" grade identically (the
// number-token matcher treats them as distinct tokens).
func tcComma(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	return out
}

// tcNumberWords are the English word forms accepted for small results,
// mirroring the grader's AnswerNumber latitude.
var tcNumberWords = []string{
	"zero", "one", "two", "three", "four", "five", "six", "seven", "eight",
	"nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen",
	"sixteen", "seventeen", "eighteen", "nineteen", "twenty",
}

// buildTempCalc generates the v7 temporal-arithmetic suite.
func buildTempCalc(r *rand.Rand, seed int64, plan *persona.Plan, n, nWaves int) tempCalcSuite {
	var out tempCalcSuite
	nCases := tempCalcCasesFor(n, nWaves)
	if nCases == 0 {
		return out
	}
	pairIdx := 0
	seedPair := func(session int, prompt, response string) {
		ts := persona.TimeAnchor(plan).Add(time.Duration(6+session)*time.Hour + time.Duration(pairIdx*beatSpacingMinutes)*time.Minute)
		out.Pairs = append(out.Pairs, protocol.MemoryPair{
			PairID:    fmt.Sprintf("p-tc-%d", pairIdx),
			SessionID: fmt.Sprintf("sess-tc-%d", session),
			Timestamp: ts.Format(time.RFC3339),
			Prompt:    prompt,
			Response:  response,
		})
		pairIdx++
	}

	domPerm := r.Perm(len(tcDomains))
	for c := 0; c < nCases; c++ {
		dom := tcDomains[domPerm[c%len(domPerm)]]
		base := dom.baseLo + dom.step*r.Intn(dom.baseN)
		delta := dom.deltaLo + dom.dstep*r.Intn(dom.deltaN)
		ans := dom.combine(base, delta)
		// The result must be distinct from both inputs, or a stored-number echo
		// would accidentally grade correct. The ranges make a collision rare; the
		// bounded redraw keeps the stream deterministic when one happens.
		for i := 0; i < 8 && (ans == base || ans == delta || ans <= 0); i++ {
			delta = dom.deltaLo + dom.dstep*r.Intn(dom.deltaN)
			ans = dom.combine(base, delta)
		}

		// The base and the change land in two NON-ADJACENT sessions so no single
		// window holds both inputs.
		seedPair(2*c, fmt.Sprintf(dom.base[r.Intn(len(dom.base))], tcComma(base)), "Noted.")
		seedPair(2*c+2, fmt.Sprintf(dom.change[r.Intn(len(dom.change))], tcComma(delta)), "Got it.")

		accept := []string{tcComma(ans)}
		if plain := fmt.Sprintf("%d", ans); plain != accept[0] {
			accept = append(accept, plain)
		}
		// Small results also accept the English word form ("nine PTO days"), the
		// same latitude grade's AnswerNumber kind gives a count.
		if ans >= 0 && ans < len(tcNumberWords) {
			accept = append(accept, tcNumberWords[ans])
		}
		out.Cases = append(out.Cases, StagedCase{Case: protocol.MemoryCase{
			ID:             protocol.OpaqueCaseID(seed, "memtc", c),
			QuestionType:   QTTempCalc,
			Question:       dom.ask[r.Intn(len(dom.ask))],
			ExpectedAnswer: accept[0],
			AnswerKind:     protocol.AnswerValue,
			AcceptAny:      accept,
		}, RunAfterWave: nWaves - 1})
	}
	return out
}
