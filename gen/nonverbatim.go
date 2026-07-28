package gen

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// bench_version 6 non-verbatim / computed answers (plan 4.3): break the
// cleartext-substring property. The load-bearing fact is stated in one unit and the
// question asks it in another, so the answer token appears in NO seeded pair — a
// literal grep fails, but a reader who converts succeeds. The answer stays
// deterministically gradeable via the accept-set (protocol.MemoryCase.AcceptAny):
// every equivalent surface form of the converted value is accepted, and CRUCIALLY
// the original stored form is NOT in the accept-set, so a parser that echoes the
// stored "90 minutes" without converting scores zero.
//
// Deterministic per (seed, draw position); v6-gated so v5's bytes are untouched.

const QTNonVerbatim = "nonverbatim-computed"

type nonVerbatimSuite struct {
	Cases []StagedCase
	Pairs []protocol.MemoryPair
}

func nonVerbatimEnabled(benchVersion int) bool { return benchVersion >= protocol.BenchVersionV6 }

// nonVerbatimCasesFor is the seed-independent case quota. Each case seeds one fact
// (stated in the un-asked unit) and asks the converted value.
func nonVerbatimCasesFor(n, nWaves int) int {
	switch {
	case nWaves < 2:
		return 0
	case n >= 100:
		return 5
	case n >= 70:
		return 3
	case n >= 40:
		return 2
	default:
		return 0
	}
}

func nonVerbatimCasesForVersion(n, nWaves, benchVersion int) int {
	if benchVersion >= protocol.BenchVersionV8 {
		switch {
		case nWaves < 2:
			return 0
		case n >= 100:
			return 30
		case n >= 40:
			return 8
		default:
			return 0
		}
	}
	return nonVerbatimCasesFor(n, nWaves)
}

// convSpec is a stated fact whose answer requires a unit conversion. stored is how
// the user states it (the un-asked unit); ask asks for the converted unit; accept is
// the set of correct converted surface forms; the stored form is deliberately absent
// from accept so an un-converted echo fails.
type convSpec struct {
	stored []string // how the user states the fact (fmt: %s = the stored quantity phrase)
	ask    []string // question surfaces
	// entries pair a concrete stored quantity phrase with its converted accept-set.
	entries []convEntry
}

type convEntry struct {
	storedQty string   // "90-minute", "a 90 minute"
	accept    []string // converted forms; NONE a substring of storedQty
}

var convSpecs = []convSpec{
	{ // minutes -> hours
		stored: []string{
			"#lead# my %s train is how I get to work.#trail#",
			"#lead# I take the %s train each way.#trail#",
		},
		ask: []string{
			"How many hours is my commute each way?",
			"In hours, how long is my train commute?",
			"My commute — how long is that in hours?",
		},
		entries: []convEntry{
			{storedQty: "30-minute", accept: []string{"0.5 hours", "0.5 hour", "half an hour", "a half hour", "0.5"}},
			{storedQty: "45-minute", accept: []string{"0.75 hours", "three quarters of an hour", "three-quarters of an hour", "0.75"}},
			{storedQty: "90-minute", accept: []string{"1.5 hours", "1.5 hour", "an hour and a half", "one and a half hours", "1.5"}},
			{storedQty: "150-minute", accept: []string{"2.5 hours", "two and a half hours", "two and a half", "2.5"}},
		},
	},
	{ // dozens -> units
		stored: []string{
			"#lead# I ordered %s of the good candles.#trail#",
			"#lead# I picked up %s of those candles.#trail#",
		},
		ask: []string{
			"How many candles did I order, as a number?",
			"In total units, how many candles did I get?",
			"What's the total count of candles I ordered?",
		},
		entries: []convEntry{
			{storedQty: "two dozen", accept: []string{"24 candles", "24 units", "24"}},
			{storedQty: "three dozen", accept: []string{"36 candles", "36 units", "36"}},
			{storedQty: "half a dozen", accept: []string{"6 candles", "6 units", "six candles", "6"}},
			{storedQty: "a dozen and a half", accept: []string{"18 candles", "18 units", "18"}},
		},
	},
	{ // weeks -> days
		stored: []string{
			"#lead# my trip is %s long.#trail#",
			"#lead# I'll be away for %s.#trail#",
		},
		ask: []string{
			"How many days is my trip?",
			"In days, how long is my trip?",
			"My trip length — how many days is that?",
		},
		entries: []convEntry{
			{storedQty: "two weeks", accept: []string{"14 days", "fourteen days", "14"}},
			{storedQty: "three weeks", accept: []string{"21 days", "twenty-one days", "21"}},
			{storedQty: "a week and a half", accept: []string{"10 days", "ten days", "10"}},
			{storedQty: "four weeks", accept: []string{"28 days", "twenty-eight days", "28"}},
		},
	},
}

// convSpecsV7Extra are ADDITIONAL non-verbatim domains available only under v7
// (appended to convSpecs by convSpecsForVersion). They must stay out of the base
// convSpecs slice: buildNonVerbatim draws r.Perm(len(specs)), so growing the
// base pool would perturb the v6 RNG stream and move v6's frozen bytes.
var convSpecsV7Extra = []convSpec{
	{ // feet -> inches
		stored: []string{
			"#lead# the bookshelf I'm building is %s tall.#trail#",
			"#lead# my new bookshelf stands %s.#trail#",
		},
		ask: []string{
			"How tall is my bookshelf in inches?",
			"In inches, how tall is the bookshelf?",
			"The bookshelf height — what's that in inches?",
		},
		entries: []convEntry{
			{storedQty: "4 feet", accept: []string{"48 inches", "forty-eight inches", "48"}},
			{storedQty: "5 feet", accept: []string{"60 inches", "sixty inches", "60"}},
			{storedQty: "six feet", accept: []string{"72 inches", "seventy-two inches", "72"}},
			{storedQty: "three and a half feet", accept: []string{"42 inches", "forty-two inches", "42"}},
		},
	},
	{ // years -> months
		stored: []string{
			"#lead# I've been at my job for %s now.#trail#",
			"#lead# my tenure at the company is %s.#trail#",
		},
		ask: []string{
			"How many months have I been at my job?",
			"In months, how long have I been at the company?",
			"My job tenure — how many months is that?",
		},
		entries: []convEntry{
			{storedQty: "two years", accept: []string{"24 months", "twenty-four months", "24"}},
			{storedQty: "a year and a half", accept: []string{"18 months", "eighteen months", "18"}},
			{storedQty: "three years", accept: []string{"36 months", "thirty-six months", "36"}},
			{storedQty: "half a year", accept: []string{"6 months", "six months", "6"}},
		},
	},
}

// convSpecsForVersion returns the base non-verbatim pool for pre-v7 contracts
// (so their RNG stream and bytes are unchanged) and the base + v7 extras for v7.
func convSpecsForVersion(benchVersion int) []convSpec {
	if benchVersion >= protocol.BenchVersionV7 {
		out := make([]convSpec, 0, len(convSpecs)+len(convSpecsV7Extra))
		out = append(out, convSpecs...)
		out = append(out, convSpecsV7Extra...)
		return out
	}
	return convSpecs
}

// buildNonVerbatim generates the v6 non-verbatim / computed-answer suite.
func buildNonVerbatim(r *rand.Rand, seed int64, plan *persona.Plan, n, nWaves, benchVersion int) nonVerbatimSuite {
	var out nonVerbatimSuite
	nCases := nonVerbatimCasesForVersion(n, nWaves, benchVersion)
	if nCases == 0 {
		return out
	}
	pairIdx := 0
	ordinal := 0
	leadG := persona.Grammar{
		"lead":  {"", "By the way,", "Quick note:", "Oh,", "For context,"},
		"trail": {"", " Anyway.", " Just noting it.", " Thought I'd mention."},
	}
	seedPair := func(session int, prompt, response string) {
		ts := persona.TimeAnchor(plan).Add(time.Duration(6+session)*time.Hour + time.Duration(pairIdx*beatSpacingMinutes)*time.Minute)
		out.Pairs = append(out.Pairs, protocol.MemoryPair{
			PairID:    fmt.Sprintf("p-nv-%d", pairIdx),
			SessionID: fmt.Sprintf("sess-nv-%d", session),
			Timestamp: ts.Format(time.RFC3339),
			Prompt:    prompt,
			Response:  response,
		})
		pairIdx++
	}

	specs := convSpecsForVersion(benchVersion)
	specPerm := r.Perm(len(specs))
	for c := 0; c < nCases; c++ {
		spec := specs[specPerm[c%len(specPerm)]]
		entry := spec.entries[r.Intn(len(spec.entries))]
		// State the fact in the un-asked unit (grammar-varied lead/trail around the
		// stored template's %s = the stored quantity phrase).
		tmpl := spec.stored[r.Intn(len(spec.stored))]
		lead := persona.Expand(r, leadG, "lead")
		trail := persona.Expand(r, leadG, "trail")
		body := fmt.Sprintf(tmpl, entry.storedQty)
		body = replaceTag(body, "#lead#", lead)
		body = replaceTag(body, "#trail#", trail)
		seedPair(c, body, "Good to know.")

		out.Cases = append(out.Cases, StagedCase{Case: protocol.MemoryCase{
			ID:             protocol.OpaqueCaseID(seed, "memnv", ordinal),
			QuestionType:   QTNonVerbatim,
			Question:       spec.ask[r.Intn(len(spec.ask))],
			ExpectedAnswer: entry.accept[0],
			AnswerKind:     protocol.AnswerValue,
			AcceptAny:      entry.accept,
		}, RunAfterWave: nWaves - 1})
		ordinal++
	}
	return out
}

// replaceTag substitutes a single literal tag, trimming a doubled leading space
// when the tag rendered empty at the start of the sentence.
func replaceTag(s, tag, val string) string {
	out := ""
	for {
		i := indexOf(s, tag)
		if i < 0 {
			out += s
			break
		}
		out += s[:i] + val
		s = s[i+len(tag):]
	}
	// tidy: leading space, doubled spaces
	out = tidySpaces(out)
	return out
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func tidySpaces(s string) string {
	// collapse runs of spaces and trim a leading space
	var b []byte
	prevSpace := true // trim leading
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			if prevSpace {
				continue
			}
			prevSpace = true
			b = append(b, ' ')
		} else {
			prevSpace = false
			b = append(b, s[i])
		}
	}
	// trim trailing space
	for len(b) > 0 && b[len(b)-1] == ' ' {
		b = b[:len(b)-1]
	}
	return string(b)
}
