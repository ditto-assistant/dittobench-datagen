package gen

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ditto-assistant/dittobench-datagen/persona"
	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

// Temporal-depth beyond latest-value (bench_version 5; v5 plan section 4.6).
// knowledge-update and point-in-time test the LATEST value (or the value as of a
// date); a grounded champion aces both. Temporal-depth asks for the
// SECOND-most-recent among three or more coexisting values:
//
//	session 1: "My favorite color is teal."
//	session 2: "I've switched my favorite color to amber."
//	session 3: "My favorite color is now <coined-blue>."
//	question:  "What was my favorite color just BEFORE I changed it to <coined-blue>?"  -> amber
//
// The answer is present in the store but is neither the recency-top pair (that is
// the latest value) nor a lexical match for the query (which names the LATEST
// value). Grep and naive recency both fail; only ordered reasoning over the
// timeline succeeds. This extends the reversal/ordering cases one level deeper and
// is a strong parser AND champion discriminator: the latest value and the oldest
// value are both distractors, so surfacing either zeros.
const (
	// QTTemporalDepth asks for the Nth-from-latest value in an update chain, not
	// the current one.
	QTTemporalDepth = "temporal-depth"
)

type temporalDepthSuite struct {
	Cases []StagedCase
	Pairs []protocol.MemoryPair
}

func temporalDepthEnabled(benchVersion int) bool { return benchVersion >= protocol.BenchVersionV5 }

// temporalDepthCasesFor is the seed-independent quota; a chain needs three seeded
// values before the question and benefits from spanning waves, so single-wave runs
// carry none.
func temporalDepthCasesFor(n, nWaves int) int {
	switch {
	case nWaves < 2:
		return 0
	case n >= 70:
		return 3
	case n >= 40:
		return 2
	case n >= 20:
		return 1
	default:
		return 0
	}
}

// tdAttr is an update-chain attribute whose values are COINED tokens so grading is
// exact and the values never collide with the haystack. The final "now" value uses
// a coined token too, named in the question, so the question cannot leak the
// answer (the answer is the PRIOR value).
type tdAttr struct {
	noun     string // "favorite color"
	changeV1 string // "My %s is %s."
	changeV2 string // "I've switched my %s to %s." (%s noun, %s value)
	changeV3 string // "My %s is now %s."
	ask      string // "What was my %s just before I changed it to %s?" (%s noun, %s v3)
}

var tdAttrs = []tdAttr{
	{noun: "favorite color", changeV1: "My %s is %s.", changeV2: "I've switched my %s to %s.", changeV3: "My %s is now %s.", ask: "What was my %s just before I changed it to %s?"},
	{noun: "go-to coffee order", changeV1: "My %s is a %s.", changeV2: "I've changed my %s to a %s.", changeV3: "My %s is now a %s.", ask: "What was my %s right before I switched it to a %s?"},
	{noun: "workout of choice", changeV1: "My %s is %s.", changeV2: "I moved my %s to %s.", changeV3: "My %s is now %s.", ask: "What was my %s just before it became %s?"},
	{noun: "primary side project", changeV1: "My %s is %s.", changeV2: "I pivoted my %s to %s.", changeV3: "My %s is now %s.", ask: "What was my %s immediately before I changed it to %s?"},
}

// buildTemporalDepth generates the temporal-depth cases. Deterministic per (seed,
// draw position); v5-gated at a fixed point in GenerateMemorySuite.
func buildTemporalDepth(r *rand.Rand, seed int64, plan *persona.Plan, n, nWaves int) temporalDepthSuite {
	var out temporalDepthSuite
	nCases := temporalDepthCasesFor(n, nWaves)
	if nCases == 0 {
		return out
	}
	pairIdx := 0
	ordinal := 0
	// Three values seeded across three distinct sessions with increasing day
	// offsets, so the timeline order is unambiguous from the timestamps.
	seedPair := func(session, dayOffset int, prompt, response string) {
		ts := persona.TimeAnchor(plan).Add(time.Duration(dayOffset)*24*time.Hour + time.Duration(session)*time.Hour + time.Duration(pairIdx*beatSpacingMinutes)*time.Minute)
		out.Pairs = append(out.Pairs, protocol.MemoryPair{
			PairID:    fmt.Sprintf("p-td-%d", pairIdx),
			SessionID: fmt.Sprintf("sess-td-%d", session),
			Timestamp: ts.Format(time.RFC3339),
			Prompt:    prompt,
			Response:  response,
		})
		pairIdx++
	}

	for c := 0; c < nCases; c++ {
		a := tdAttrs[c%len(tdAttrs)]
		v1 := persona.CoinShaped(seed, fmt.Sprintf("td|v1|%d", c))
		v2 := persona.CoinShaped(seed, fmt.Sprintf("td|v2|%d", c)) // the ANSWER (second-most-recent)
		v3 := persona.CoinShaped(seed, fmt.Sprintf("td|v3|%d", c)) // the latest (named in the question)

		seedPair(3*c, c, fmt.Sprintf(a.changeV1, a.noun, v1), "Noted.")
		seedPair(3*c+1, c+2, fmt.Sprintf(a.changeV2, a.noun, v2), "Got it.")
		seedPair(3*c+2, c+5, fmt.Sprintf(a.changeV3, a.noun, v3), "Updated.")

		mc := protocol.MemoryCase{
			ID:             protocol.OpaqueCaseID(seed, "memtd", ordinal),
			QuestionType:   QTTemporalDepth,
			Question:       fmt.Sprintf(a.ask, a.noun, v3),
			ExpectedAnswer: v2,
			AnswerKind:     protocol.AnswerValue,
			// The latest value (v3) and the oldest (v1) are both wrong: surfacing the
			// current value (naive recency) or the original (grep the earliest) zeros.
			DistractorAnswers: []string{v3, v1},
		}
		out.Cases = append(out.Cases, StagedCase{Case: mc, RunAfterWave: nWaves - 1})
		ordinal++
	}
	return out
}
